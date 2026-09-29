import asyncio
from typing import AsyncGenerator
from worker.inference.clients.base import BaseLLMProvider
from worker.inference.clients.mistral import mistral_provider
from worker.inference.streaming import stream_handler
from worker.inference.inference import inference_pb2


class InferenceChain:
    """
    Async inference chain coordinating client execution and streaming output.

    Accepts any BaseLLMProvider so the chain is provider-agnostic — swap in an
    OpenAIProvider, AnthropicProvider, etc. without touching this class.
    Uses astream_chat() and StreamHandler so all concurrent requests share one
    asyncio event loop instead of blocking OS threads.
    """

    def __init__(self, provider: BaseLLMProvider = mistral_provider):
        self.client = provider
        self.stream_handler = stream_handler

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

        astream_gen = self.client.astream_chat(  # yields (content, prompt_toks, completion_toks)
            messages=lc_messages,
            model=model,
            temperature=temp,
            max_tokens=max_toks,
        )

        async for resp in self.stream_handler.stream_inference(
            request=request,
            astream_gen=astream_gen,
            estimated_prompt_tokens=estimated_prompt_tokens,
        ):
            yield resp


inference_chain = InferenceChain()
