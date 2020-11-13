FROM github/super-linter:latest

RUN pip3 install pycodestyle bandit
RUN apk add --no-cache go git

ENV GO111MODULE="on"

RUN go get -v github.com/cugu/dashboard@master github.com/eth0izzle/shhgit

ADD https://raw.githubusercontent.com/eth0izzle/shhgit/master/config.yaml /shhgit/config.yaml 

ENTRYPOINT [""]
