import asyncio
from typing import List, AsyncGenerator, Tuple
from langchain_core.messages import BaseMessage, HumanMessage, AIMessage, SystemMessage
from worker.inference.models.config import config


class MistralClient:
    """
    Async-first LangChain/Mistral client.

    Uses LangChain's astream() so the inference worker runs on a single asyncio
    event loop (via grpc.aio) rather than one OS thread per concurrent stream.
    This allows hundreds of concurrent streaming RPCs without spawning hundreds
    of threads — the bottleneck becomes Mistral API concurrency, not thread count.
    """

    def __init__(self):
        self.api_key = config.mistral_api_key
        self._llm_cache: dict = {} # cache_key -> ChatMistralAI

    def _get_llm(self, model: str, temperature: float, max_tokens: int):
        from langchain_mistralai import ChatMistralAI
        cache_key = f"{model}_{temperature}_{max_tokens}"
        if cache_key not in self._llm_cache:
            self._llm_cache[cache_key] = ChatMistralAI(
                model=model,
                temperature=temperature,
                max_tokens=max_tokens if max_tokens > 0 else None,
                mistral_api_key=self.api_key,
            )
        return self._llm_cache[cache_key]

    def convert_messages(self, proto_messages) -> List[BaseMessage]:
        messages: List[BaseMessage] = []
        for msg in proto_messages:
            role = msg.role.lower()
            if role in ("system", "developer"):
                messages.append(SystemMessage(content=msg.content))
            elif role == "user":
                messages.append(HumanMessage(content=msg.content))
            elif role == "assistant":
                messages.append(AIMessage(content=msg.content))
            else:
                messages.append(HumanMessage(content=msg.content))
        return messages

    async def astream_chat(
        self,
        messages: List[BaseMessage],
        model: str,
        temperature: float = 0.7,
        max_tokens: int = 500,
    ) -> AsyncGenerator[Tuple[str, int, int], None]:
        """
        Async streaming from Mistral via LangChain astream().

        Yields (chunk_text, prompt_tokens, completion_tokens).
        prompt/completion tokens are 0 for intermediate chunks and populated
        in the final chunk via usage_metadata when available.

        Falls back to a deterministic async mock when MISTRAL_API_KEY is unset.
        """
        if self.api_key:
            llm = self._get_llm(model, temperature, max_tokens)
            async for chunk in llm.astream(messages):
                content = chunk.content
                usage = getattr(chunk, "usage_metadata", None)
                prompt_toks = usage.get("input_tokens", 0) if usage else 0
                completion_toks = usage.get("output_tokens", 0) if usage else 0

                if isinstance(content, str) and content:
                    yield (content, prompt_toks, completion_toks)
                elif isinstance(content, list):
                    text_parts = "".join(str(p) for p in content if p)
                    if text_parts:
                        yield (text_parts, prompt_toks, completion_toks)
        else:
            # Async mock: simulates token-by-token delay without blocking the event loop
            prompt_summary = " ".join(
                m.content for m in messages if isinstance(m, HumanMessage)
            )
            
            mock_text = (
                f"Aegis async mock response for model '{model}'. "
                f"Query received: {prompt_summary}"
            )
            tokens = mock_text.split(" ")
            for i, token in enumerate(tokens):
                await asyncio.sleep(0.01)  # non-blocking delay; yields to event loop
                suffix = " " if i < len(tokens) - 1 else ""
                yield (token + suffix, 0, 0)


mistral_client = MistralClient()
