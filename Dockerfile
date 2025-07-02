# syntax=docker/dockerfile:1

FROM alpine:3.19

LABEL maintainer="zack"
LABEL description="Simple test container for SYAC image builds"

# Create a dummy application
RUN echo -e '#!/bin/sh\\necho \"Hello from SYAC test image!\"' > /hello.sh && \
    chmod +x /hello.sh

ENTRYPOINT ["/hello.sh"]