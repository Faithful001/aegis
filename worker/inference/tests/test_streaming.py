import unittest
import asyncio
from unittest.mock import AsyncMock, patch
from worker.inference.clients.mistral_client import mistral_client
from worker.inference.chains.inference_chain import inference_chain
from worker.inference.inference import inference_pb2


class TestMistralClientStreaming(unittest.IsolatedAsyncioTestCase):
    async def test_astream_chat_mock_fallback(self):
        """Verify mock streaming when no Mistral API key is provided."""
        from langchain_core.messages import HumanMessage
        messages = [HumanMessage(content="Hello streaming test")]
        
        chunks = []
        async for chunk_text, prompt_toks, comp_toks in mistral_client.astream_chat(
            messages=messages,
            model="mistral-small",
            temperature=0.7,
            max_tokens=50,
        ):
            chunks.append(chunk_text)
            
        full_text = "".join(chunks)
        self.assertIn("Aegis async mock response", full_text)
        self.assertIn("mistral-small", full_text)


class TestInferenceChainStreaming(unittest.IsolatedAsyncioTestCase):
    async def test_execute_stream_true(self):
        """Test streaming execution yields intermediate chunks and final usage chunk."""
        proto_messages = [
            inference_pb2.ChatMessage(role="user", content="Explain quantum computing")
        ]
        request = inference_pb2.GenerateRequest(
            request_id="req_stream_test_01",
            job_id="job_01",
            model="mistral-small",
            messages=proto_messages,
            max_tokens=100,
            temperature=0.7,
            stream=True,
            organization_id="org_123",
            project_id="proj_456",
        )

        responses = []
        async for resp in inference_chain.execute(request):
            responses.append(resp)

        self.assertGreater(len(responses), 1)
        
        # Intermediate chunks should have done=False
        for intermediate in responses[:-1]:
            self.assertEqual(intermediate.request_id, "req_stream_test_01")
            self.assertFalse(intermediate.done)
            
        # Final chunk should have done=True and usage metadata
        final = responses[-1]
        self.assertTrue(final.done)
        self.assertEqual(final.finish_reason, "stop")
        self.assertIsNotNone(final.usage)
        self.assertGreater(final.usage.total_tokens, 0)

    async def test_execute_stream_false(self):
        """Test unary execution returns single done response with accumulated content."""
        proto_messages = [
            inference_pb2.ChatMessage(role="user", content="Unary test")
        ]
        request = inference_pb2.GenerateRequest(
            request_id="req_unary_test_01",
            job_id="job_02",
            model="mistral-small",
            messages=proto_messages,
            max_tokens=50,
            temperature=0.5,
            stream=False,
        )

        responses = []
        async for resp in inference_chain.execute(request):
            responses.append(resp)

        self.assertEqual(len(responses), 1)
        resp = responses[0]
        self.assertTrue(resp.done)
        self.assertEqual(resp.finish_reason, "stop")
        self.assertIn("Aegis async mock response", resp.content)
        self.assertIsNotNone(resp.usage)

    async def test_execute_cancellation_handling(self):
        """Test cancellation handling yields cancelled finish_reason."""
        proto_messages = [
            inference_pb2.ChatMessage(role="user", content="Cancel me")
        ]
        request = inference_pb2.GenerateRequest(
            request_id="req_cancel_01",
            job_id="job_03",
            model="mistral-small",
            messages=proto_messages,
            stream=True,
        )

        async def mock_cancelled_astream(*args, **kwargs):
            yield ("Token 1 ", 0, 0)
            raise asyncio.CancelledError()

        with patch.object(mistral_client, "astream_chat", side_effect=mock_cancelled_astream):
            responses = []
            async for resp in inference_chain.execute(request):
                responses.append(resp)

            self.assertEqual(len(responses), 2)
            self.assertFalse(responses[0].done)
            self.assertEqual(responses[0].content, "Token 1 ")
            
            final = responses[1]
            self.assertTrue(final.done)
            self.assertEqual(final.finish_reason, "cancelled")
            self.assertEqual(final.error, "request cancelled")


if __name__ == "__main__":
    unittest.main()
