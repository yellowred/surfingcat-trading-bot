FROM node:9

RUN apt-get -yqq update && \
    apt-get -y install net-tools && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

COPY package.json /tmp/package.json

RUN \
    cd /tmp && \
    npm install --production && \
    npm cache verify

RUN \
  mkdir /app && \
  mkdir /app/log && \
  cp -a /tmp/node_modules/ /app

ADD ./ /app

WORKDIR /app

EXPOSE 8080

CMD npm run start