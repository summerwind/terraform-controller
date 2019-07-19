FROM ubuntu:18.04 as base

ENV TERRAFORM_VERSION=0.12.5

RUN apt update && apt install -y ca-certificates unzip curl gnupg

RUN curl -L -O https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip \
  && curl -L -O https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_SHA256SUMS \
  && curl -L -O https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_SHA256SUMS.sig \
  && curl https://keybase.io/hashicorp/pgp_keys.asc | gpg --import \
  && gpg --verify terraform_${TERRAFORM_VERSION}_SHA256SUMS.sig terraform_${TERRAFORM_VERSION}_SHA256SUMS \
  && grep linux_amd64 terraform_${TERRAFORM_VERSION}_SHA256SUMS >terraform_${TERRAFORM_VERSION}_SHA256SUMS_linux_amd64 \
  && sha256sum -c --status terraform_${TERRAFORM_VERSION}_SHA256SUMS_linux_amd64 \
  && unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d /bin

#######################################

FROM golang:1.12 as builder

ENV GO111MODULE=on \
    GOPROXY=https://proxy.golang.org

WORKDIR /go/src/github.com/summerwind/terraform-controller
COPY go.mod go.sum .
RUN go mod download

WORKDIR /workspace
COPY . .

RUN go vet ./...
RUN go test -v ./...
RUN CGO_ENABLED=0 go build .

#######################################

FROM summerwind/whitebox-controller:latest

COPY --from=base /tmp /tmp
COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=base /bin/terraform /bin/terraform

COPY --from=builder /workspace/terraform-controller /bin/terraform-controller
COPY --from=builder /workspace/config.yaml /config.yaml

ENTRYPOINT ["/bin/whitebox-controller"]
