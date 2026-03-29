FROM golang:1.26

ARG USERNAME=gopher

RUN groupadd --gid 1000 $USERNAME \
  && useradd --uid 1000 --gid $USERNAME --shell /bin/bash --create-home $USERNAME
