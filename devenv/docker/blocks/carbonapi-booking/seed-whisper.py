"""
Create minimal seed whisper files so go-carbon's carbonserver has at least
one metric in its index on startup.  Without this, go-carbon returns HTTP 500
for find queries on an empty index, which bookingcom/carbonapi propagates as
HTTP 422 (Unprocessable Entity) rather than treating it as "not found".

We create two metrics:
  - test.integration  (used by the go-carbon healthcheck: query=test.*)
  - grafanatest       (flat metric used by the integration test suite)
"""
import struct, os

def write_wsp(path):
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "wb") as f:
        # Whisper header: aggregation=average(1), maxRetention=86400s, xff=0.5, archives=1
        f.write(struct.pack(">LLfL", 1, 86400, 0.5, 1))
        # Archive 0: data at byte 28 (16 header + 12 archive-info), 60s/pt, 1440 points
        f.write(struct.pack(">LLL", 28, 60, 1440))
        # 1440 empty data points (timestamp=0, value=0.0)
        f.write(b"\x00" * (1440 * 12))

write_wsp("/var/lib/go-carbon/whisper/test/integration.wsp")
write_wsp("/var/lib/go-carbon/whisper/grafanatest.wsp")
