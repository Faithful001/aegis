from abc import ABC, abstractmethod
from typing import AsyncGenerator, List, Tuple

from langchain_core.messages import BaseMessage


class BaseLLMProvider(ABC):
    """
    Abstract contract for all LLM provider adapters in the Aegis inference worker.

    Any concrete implementation (e.g. MistralProvider, OpenAIProvider) must
    implement both methods so that InferenceChain can remain provider-agnostic.
    """

    @abstractmethod
    async def astream_chat(
        self,
        messages: List[BaseMessage],
        model: str,
        temperature: float,
        max_tokens: int,
    ) -> AsyncGenerator[Tuple[str, int, int], None]:
        """
        Stream a chat completion token-by-token.

        Yields:
            (chunk_text, prompt_tokens, completion_tokens) tuples.
            prompt_tokens and completion_tokens are 0 for intermediate chunks
            and populated (when available from the provider) in the final chunk.
        """
        ...

    @abstractmethod
    def convert_messages(self, proto_messages) -> List[BaseMessage]:
        """
        Convert a list of proto ChatMessage objects into LangChain BaseMessages.

        Args:
            proto_messages: Iterable of inference_pb2.ChatMessage instances.

        Returns:
            A list of LangChain BaseMessage objects ready for the LLM.
        """
        ...
