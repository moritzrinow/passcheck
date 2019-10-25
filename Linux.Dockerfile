FROM golang AS build
LABEL maintainer="Moritz Rinow <mrinow.dev@gmail.com>"
WORKDIR /src
COPY . .
RUN make

FROM alpine AS final
WORKDIR /usr/bin/passcheck
COPY --from=build /src/bin .
RUN apk add --no-cache bash
RUN echo "export PATH=$PATH:." >> ~/.bashrc
CMD ["/bin/bash"]