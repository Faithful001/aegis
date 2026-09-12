import asyncio
from typing import AsyncGenerator
from worker.inference.clients.mistral_client import mistral_client
from worker.inference.inference import inference_pb2


class InferenceChain:
    """
    Async inference chain.

    Uses astream_chat() so all concurrent requests share one asyncio event loop
    instead of blocking one OS thread each.
    """

    def __init__(self):
        self.client = mistral_client

    async def execute(
        self,
        request: inference_pb2.GenerateRequest,
    ) -> AsyncGenerator[inference_pb2.GenerateResponse, None]:
        lc_messages = self.client.convert_messages(request.messages)
        model = request.model or "mistral-small-latest"
        temp = request.temperature if request.temperature > 0 else 0.7
        max_toks = request.max_tokens if request.max_tokens > 0 else 500

        # Rough prompt token estimate (4 chars ≈ 1 token)
        total_prompt_chars = sum(len(m.content) for m in lc_messages)
        estimated_prompt_tokens = max(1, total_prompt_chars // 4)

        output_tokens = 0
        full_content = []
        last_prompt_toks = estimated_prompt_tokens
        last_completion_toks = 0

        try:
            async for chunk_text, prompt_toks, completion_toks in self.client.astream_chat(
                messages=lc_messages,
                model=model,
                temperature=temp,
                max_tokens=max_toks,
            ):
                output_tokens += 1
                full_content.append(chunk_text)

                # LangChain's usage_metadata populates on the final chunk
                if prompt_toks > 0:
                    last_prompt_toks = prompt_toks
                if completion_toks > 0:
                    last_completion_toks = completion_toks

                if request.stream:
                    yield inference_pb2.GenerateResponse(
                        request_id=request.request_id,
                        job_id=request.job_id,
                        content=chunk_text,
                        done=False,
                        finish_reason="",
                        usage=None,
                        error="",
                    )

            # Use real token counts if available, otherwise fall back to estimate
            final_completion = last_completion_toks if last_completion_toks > 0 else output_tokens
            final_prompt = last_prompt_toks
            final_usage = inference_pb2.TokenUsage(
                prompt_tokens=final_prompt,
                completion_tokens=final_completion,
                total_tokens=final_prompt + final_completion,
            )

            if request.stream:
                yield inference_pb2.GenerateResponse(
                    request_id=request.request_id,
                    job_id=request.job_id,
                    content="",
                    done=True,
                    finish_reason="stop",
                    usage=final_usage,
                    error="",
                )
            else:
                yield inference_pb2.GenerateResponse(
                    request_id=request.request_id,
                    job_id=request.job_id,
                    content="".join(full_content),
                    done=True,
                    finish_reason="stop",
                    usage=final_usage,
                    error="",
                )

        except asyncio.CancelledError:
            yield inference_pb2.GenerateResponse(
                request_id=request.request_id,
                job_id=request.job_id,
                content="",
                done=True,
                finish_reason="cancelled",
                usage=None,
                error="request cancelled",
            )
        except Exception as e:
            yield inference_pb2.GenerateResponse(
                request_id=request.request_id,
                job_id=request.job_id,
                content="",
                done=True,
                finish_reason="error",
                usage=None,
                error=str(e),
            )


inference_chain = InferenceChain()
