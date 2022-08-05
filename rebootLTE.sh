#!/bin/bash
curl --interface 192.168.$1.2 http://192.168.8.1/api/webserver/SesTokInfo > tokens.xml
SesInfo=$(xmllint --xpath '//SesInfo/text()' tokens.xml)
TokInfo=$(xmllint --xpath '//TokInfo/text()' tokens.xml)
rm tokens.xml

curl --interface 192.168.$1.2 -X POST http://192.168.8.1/api/device/control \
    -H "Cookie: $SesInfo" \
    -H "__RequestVerificationToken: $TokInfo" \
    -H "Content-Type: application/xml" -d "<request><Control>1</Control></request>" > result.xml

result=$(xmllint --xpath '//response/text()' result.xml)
rm result.xml
if [[ "$result" == "OK" ]]; then
    echo "successful"
else
    echo "error"
fi