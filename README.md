# Surfingcat Trading Bot in Golang for Cryptocurrencies
Welcome!

## Running Docker
1. Create new docker machine `docker-machine create -d virtualbox --virtualbox-hostonly-cidr 192.168.10.1/24 --virtualbox-memory '1024' --virtualbox-boot2docker-url https://releases.rancher.com/os/latest/rancheros.iso --engine-install-url https://raw.githubusercontent.com/SvenDowideit/install-docker/5896b863698967df0738976d6ee98efc5d4637ae/1.12.6.sh spa-sandbox`
2. `docker build -t surfingcat-trading-bot .`
3. `docker run -p 3026:3026 -d --name surfingcat-trading-bot surfingcat-trading-bot`
4. `curl "http://192.168.33.100:3026/indicator?name=ema&market=USDT-BTC&interval=50"`
