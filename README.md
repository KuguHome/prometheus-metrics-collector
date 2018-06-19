# dayreport-send
This is a program currently installed on our central component server (zentralkomponente, zk). It is running automatically once a day. It logs into the control units (steuerzentrale, SZ, RaspberryPI), downloads the tagesbericht (daily status reports) pdf files (see https://gitlab.kugu-home.com/projekte/metrics-collector/blob/master/examples/tagesbericht_2018-06-18.pdf) and sends them out to e.g. Leo & Christopher.

## how to compile

1. ```GOPATH="$GOPATH:`pwd`" go get -d ...```
1. ```GOPATH="$GOPATH:`pwd`" go install tagesbericht-versand```
