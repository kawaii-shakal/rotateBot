#!/bin/bash
curl -m 10 --interface 192.168.$1.2 http://192.168.8.1/api/webserver/SesTokInfo > tokens$1.xml
SesInfo=`xmllint --xpath '//SesInfo/text()' tokens$1.xml`
TokInfo=`xmllint --xpath '//TokInfo/text()' tokens$1.xml`
rm tokens$1.xml

curl --interface 192.168.$1.2 -X POST http://192.168.8.1/api/device/control \
    -H "Cookie: $SesInfo" \
    -H "__RequestVerificationToken: $TokInfo" \
    -H "Content-Type: application/xml" -d "<request><Control>1</Control></request>" > result$1.xml

result=`xmllint --xpath '//response/text()' result$1.xml`
rm result$1.xml
if [[ "$result" == "OK" ]]; then
    echo "successful"
else
    echo "error"
fi