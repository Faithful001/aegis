import asyncio
from typing import AsyncGenerator, Tuple, List
from worker.inference.inference import inference_pb2


class StreamHandler:
    """
    Handles formatting and lifecycle management for gRPC GenerateResponse streams.

    Provides utility methods for constructing intermediate token chunks, final completion
    chunks with usage metadata, and cancellation/error responses.
    """

    @staticmethod
    def build_chunk_response(request_id: str, job_id: str, content: str) -> inference_pb2.GenerateResponse:
        """Construct an intermediate streaming token response chunk."""
        return inference_pb2.GenerateResponse(
            request_id=request_id,
            job_id=job_id,
            content=content,
            done=False,
            finish_reason="",
            usage=None,
            error="",
        )

    @staticmethod
    def build_final_response(
        request_id: str,
        job_id: str,
        content: str,
        finish_reason: str,
        prompt_tokens: int,
        completion_tokens: int,
    ) -> inference_pb2.GenerateResponse:
        """Construct the final completion chunk carrying token usage statistics."""
        usage = inference_pb2.TokenUsage(
            prompt_tokens=prompt_tokens,
            completion_tokens=completion_tokens,
            total_tokens=prompt_tokens + completion_tokens,
        )
        return inference_pb2.GenerateResponse(
            request_id=request_id,
            job_id=job_id,
            content=content,
            done=True,
            finish_reason=finish_reason,
            usage=usage,
            error="",
        )

    @staticmethod
    def build_error_response(
        request_id: str,
        job_id: str,
        finish_reason: str,
        error_message: str,
    ) -> inference_pb2.GenerateResponse:
        """Construct a termination chunk for cancelled or failed inference streams."""
        return inference_pb2.GenerateResponse(
            request_id=request_id,
            job_id=job_id,
            content="",
            done=True,
            finish_reason=finish_reason,
            usage=None,
            error=error_message,
        )

    async def stream_inference(
        self,
        request: inference_pb2.GenerateRequest,
        astream_gen: AsyncGenerator[Tuple[str, int, int], None],
        estimated_prompt_tokens: int,
    ) -> AsyncGenerator[inference_pb2.GenerateResponse, None]:
        """
        Consumes raw token tuple stream from MistralClient and yields formatted
        gRPC GenerateResponse messages over time.
        """
        output_tokens = 0
        full_content: List[str] = []
        last_prompt_toks = estimated_prompt_tokens
        last_completion_toks = 0

        try:
            async for chunk_text, prompt_toks, completion_toks in astream_gen:
                output_tokens += 1
                full_content.append(chunk_text)

                if prompt_toks > 0:
                    last_prompt_toks = prompt_toks
                if completion_toks > 0:
                    last_completion_toks = completion_toks

                if request.stream:
                    yield self.build_chunk_response(
                        request_id=request.request_id,
                        job_id=request.job_id,
                        content=chunk_text,
                    )

            final_completion = last_completion_toks if last_completion_toks > 0 else output_tokens
            final_prompt = last_prompt_toks
            final_text = "" if request.stream else "".join(full_content)

            yield self.build_final_response(
                request_id=request.request_id,
                job_id=request.job_id,
                content=final_text,
                finish_reason="stop",
                prompt_tokens=final_prompt,
                completion_tokens=final_completion,
            )

        except asyncio.CancelledError:
            yield self.build_error_response(
                request_id=request.request_id,
                job_id=request.job_id,
                finish_reason="cancelled",
                error_message="request cancelled",
            )
        except Exception as e:
            yield self.build_error_response(
                request_id=request.request_id,
                job_id=request.job_id,
                finish_reason="error",
                error_message=str(e),
            )


stream_handler = StreamHandler()
