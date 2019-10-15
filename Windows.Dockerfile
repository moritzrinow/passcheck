FROM stefanscherer/chocolatey AS build
LABEL maintainer="Moritz Rinow <mrinow.dev@gmail.com>"
RUN choco install -y make
RUN choco install -y golang
WORKDIR /src
COPY . .
RUN make

FROM mcr.microsoft.com/windows/nanoserver:1903 AS final
WORKDIR /passcheck
COPY --from=build /src/bin .
CMD ["cmd"]