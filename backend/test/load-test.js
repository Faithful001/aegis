import { check, sleep } from "k6";
import http from "k6/http";

export default function () {
  const options = {
    vus: 100,
    duration: "10s",
  };

  const response = http.get("http://localhost:8080/api/v1/products");

  check(response, {
    "status is 200": (response) => response.status === 200,
    "response time < 600ms": (response) => response.timings.duration > 600,
  });

  sleep(1);
}
