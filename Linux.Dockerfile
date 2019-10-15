FROM golang AS build
LABEL maintainer="Moritz Rinow <mrinow.dev@gmail.com>"
WORKDIR /src
COPY . .
RUN make

FROM alpine AS final
WORKDIR /usr/bin/passcheck
COPY --from=build /src/bin .
CMD ["/bin/sh"]