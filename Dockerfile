FROM github/super-linter:latest

RUN pip3 install pycodestyle
RUN apk add --no-cache go git

ENV GO111MODULE="on"

RUN go get -v github.com/cugu/dashboard

ENTRYPOINT [""]
