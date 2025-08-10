#!/bin/bash

# 檢查 openssl 是否已安裝
if ! [ -x "$(command -v openssl)" ]; then
  echo '錯誤：需要安裝 openssl。' >&2
  exit 1
fi

# 設定憑證參數
COUNTRY="TW"
STATE="Taiwan"
LOCALITY="Taipei"
ORGANIZATION="VoIP Demo"
COMMON_NAME="localhost"
DAYS_VALID=365

# 產生私鑰和憑證
openssl req -x509 -newkey rsa:2048 -nodes \
  -keyout key.pem -out cert.pem \
  -days ${DAYS_VALID} \
  -subj "/C=${COUNTRY}/ST=${STATE}/L=${LOCALITY}/O=${ORGANIZATION}/CN=${COMMON_NAME}"

echo "憑證 'cert.pem' 和私鑰 'key.pem' 已成功產生。"
