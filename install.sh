command -v go >/dev/null 2>&1 || { echo >&2 "Golang is required for this. Aborting."; exit 1; }

go build -o restream .
cp restream /usr/local/bin/

mkdir -p /etc/restream

if [ ! -f /etc/restream/config.toml ]; then
    cp ./config.toml /etc/restream/config.toml
fi

cp ./restream.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable restream.service
systemctl start restream.service

echo "Restream service has been successfully installed and started."
