#!/usr/bin/env python3
"""
Integration test script for Phase 5 (Streaming).

Spawns the Python gRPC Inference Worker, sends a gRPC GenerateRequest with stream=True,
and verifies real-time token streaming, token usage metadata, and completion flags.
"""

import sys
import os
import asyncio
import time
import subprocess

# Add project root to sys.path
PROJECT_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
if PROJECT_ROOT not in sys.path:
    sys.path.insert(0, PROJECT_ROOT)

import grpc
from worker.inference.inference import inference_pb2, inference_pb2_grpc


async def run_streaming_integration_test():
    port = 50055
    server_address = f"localhost:{port}"

    print(f"[Phase 5 Test] Starting Python gRPC inference worker on {server_address}...")
    env = os.environ.copy()
    env["GRPC_PORT"] = str(port)

    # Use virtual environment python if available
    venv_python = os.path.join(PROJECT_ROOT, "worker", ".venv", "Scripts", "python.exe")
    if not os.path.exists(venv_python):
        venv_python = sys.executable

    worker_proc = subprocess.Popen(
        [venv_python, "-m", "worker.inference.main"],
        cwd=PROJECT_ROOT,
        env=env,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )

    try:
        # Give server a moment to bind and start listening
        await asyncio.sleep(2.0)

        # Check if worker process exited prematurely
        if worker_proc.poll() is not None:
            stdout, stderr = worker_proc.communicate()
            raise RuntimeError(f"Worker process exited with code {worker_proc.returncode}.\nSTDOUT: {stdout.decode()}\nSTDERR: {stderr.decode()}")

        async with grpc.aio.insecure_channel(server_address) as channel:
            stub = inference_pb2_grpc.InferenceServiceStub(channel)

            # 1. Health check verification
            health_req = inference_pb2.HealthCheckRequest(worker_id="test_runner")
            health_resp = await stub.HealthCheck(health_req)
            print(f"[Phase 5 Test] Worker Status: {health_resp.status}, Models: {health_resp.supported_models}")
            assert health_resp.status == "SERVING", f"Expected SERVING, got {health_resp.status}"

            # 2. Streaming Generate Request
            req = inference_pb2.GenerateRequest(
                request_id="req_phase5_e2e_001",
                job_id="job_phase5_001",
                model="mistral-small",
                messages=[
                    inference_pb2.ChatMessage(role="system", content="You are a helpful assistant."),
                    inference_pb2.ChatMessage(role="user", content="Explain gRPC streaming in 2 sentences."),
                ],
                max_tokens=100,
                temperature=0.7,
                stream=True,
                organization_id="org_test",
                project_id="proj_test",
            )

            print("[Phase 5 Test] Sending streaming Generate request over gRPC...")
            start_time = time.time()
            chunk_count = 0
            accumulated_text = ""
            final_usage = None
            finished_done = False

            stream = stub.Generate(req)
            async for resp in stream:
                chunk_count += 1
                if resp.content:
                    accumulated_text += resp.content
                    print(f"  [Chunk {chunk_count}] Content: '{resp.content}'")
                
                if resp.done:
                    finished_done = True
                    final_usage = resp.usage
                    print(f"  [Final Chunk] Done: True, Reason: {resp.finish_reason}")
                    if final_usage:
                        print(f"  [Usage] Prompt Tokens: {final_usage.prompt_tokens}, Completion Tokens: {final_usage.completion_tokens}, Total: {final_usage.total_tokens}")

            elapsed = time.time() - start_time
            print(f"[Phase 5 Test] Streaming completed in {elapsed:.3f}s across {chunk_count} chunks.")
            print(f"[Phase 5 Test] Accumulated text: '{accumulated_text.strip()}'")

            assert finished_done, "Stream did not yield a final done=True chunk"
            assert chunk_count >= 2, f"Expected at least 2 chunks for streaming, got {chunk_count}"
            assert len(accumulated_text) > 0, "Expected non-empty accumulated content"
            assert final_usage is not None, "Expected usage metadata in final chunk"
            assert final_usage.total_tokens > 0, "Expected positive total_tokens"

            print("[Phase 5 Test] SUCCESS: All streaming integration assertions passed!")

    finally:
        worker_proc.terminate()
        try:
            worker_proc.wait(timeout=2)
        except subprocess.TimeoutExpired:
            worker_proc.kill()


if __name__ == "__main__":
    asyncio.run(run_streaming_integration_test())
