FROM ghcr.io/horahoradev/liblcf:master_liblcf

COPY ynoclient ynoclient

# Build ynoclient
RUN --mount=type=cache,target=/workdir/ynoclient/build /bin/bash -c 'source buildscripts/emscripten/emsdk-portable/emsdk_env.sh && \
	ln -s /workdir /root/workdir && \
	cd ynoclient && \
	./cmake_build.sh && cd build && \
	/usr/bin/ninja && \
	echo "done"'


RUN --mount=type=cache,target=/workdir/ynoclient/build cp /workdir/ynoclient/build/index.wasm /workdir/ynoclient/
RUN --mount=type=cache,target=/workdir/ynoclient/build cp /workdir/ynoclient/build/index.js /workdir/ynoclient/

FROM ubuntu:rolling

WORKDIR /multi_server

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && \
	apt-get install -y git wget gcc && \
	mkdir public

RUN cd /usr/local && \
	wget https://golang.org/dl/go1.17.3.linux-amd64.tar.gz && \
	rm -rf /usr/local/go && \
	tar -C /usr/local -xzf go1.17.3.linux-amd64.tar.gz

ENV PATH=$PATH:/usr/local/go/bin

COPY server orbs

RUN cd orbs && \
	go mod vendor && \
    go build --mod=vendor -o /multi_server/multi_server .

RUN apt-get install -y python3 unzip python3-pip locales locales-all && \
	pip install gdown

RUN locale-gen ja_JP.UTF-8
ENV LANG ja_JP.UTF-8
ENV LANGUAGE ja_JP
ENV LC_ALL ja_JP.UTF-8

RUN gdown https://drive.google.com/uc?id=1c8g2XBLFQ6L6KNrmI3njhgk714uX0p3W -O ./public/y2kki.zip && \
	cd public && \
	unzip -O shift-jis ./y2kki.zip && \
	mkdir -p /multi_server/public/play/games/ && \
	/bin/bash -c 'mv /multi_server/public/ゆめ2っきver0.117g /multi_server/public/play/games/default'

COPY gencache /multi_server/public/play/games/default/ゆめ2っき/

RUN mv /multi_server/public/play/games/default/ゆめ2っき/music /multi_server/public/play/games/default/ゆめ2っき/Music

RUN cd /multi_server/public/play/games/default/ゆめ2っき/ && \
	./gencache

RUN /bin/bash -c 'mv /multi_server/public/play/games/default/ゆめ2っき/* /multi_server/public/play/games/default/'

COPY --from=0 /workdir/ynoclient/index.wasm /multi_server/public
COPY --from=0 /workdir/ynoclient/index.js /multi_server/public

COPY server/public /multi_server/public

RUN mkdir -p /multi_server/public/data/default && \
	cp -r /multi_server/public/play/games/default/* /multi_server/public/data/default

ENTRYPOINT ["./multi_server"]
