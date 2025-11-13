FROM ubuntu:20.04

# 安装基本依赖
RUN apt-get update && apt-get install -y wget && apt-get clean

# 更新apt源
RUN wget -O /etc/apt/sources.list https://mirrors.tencent.com/repo/ubuntu20_sources.list

# 提供ARG用于检测当前架构
ARG TARGETARCH

RUN if [ "$TARGETARCH" = "arm64" ]; then \
      sed -i 's#/ubuntu/#/ubuntu-ports/#' /etc/apt/sources.list; \
    fi

# 安装基本依赖
RUN apt-get update && apt-get install -y \
    python3 \
    python3-pip \
    python3-dev \
    patchelf \
    build-essential \
    autoconf \
    automake \
    libffi-dev \
    libssl-dev \
    git \
    wget \
    scons \
    curl \
    && apt-get clean

WORKDIR /app

RUN pip config set global.index-url https://mirrors.tencent.com/pypi/simple/

# 安装PyInstaller和StaticX
RUN pip3 install --upgrade pip
RUN pip3 install setuptools wheel
RUN pip3 install pyinstaller
RUN pip3 install staticx

# 安装Python依赖
RUN pip3 install jinja2
RUN pip3 install six

COPY jinja2_exec.py /app/

# 使用PyInstaller构建可执行文件
RUN pyinstaller --onefile --clean --strip --noupx jinja2_exec.py

RUN if [ "$TARGETARCH" = "amd64" ]; then \
        staticx --strip dist/jinja2_exec dist/jinja2_exec_static \
            -l /lib/x86_64-linux-gnu/libnss_dns.so.2 \
            -l /lib/x86_64-linux-gnu/libresolv.so.2 \
            -l /lib/x86_64-linux-gnu/libgcc_s.so.1; \
    elif [ "$TARGETARCH" = "arm64" ]; then \
        staticx --strip dist/jinja2_exec dist/jinja2_exec_static \
            -l /lib/aarch64-linux-gnu/libnss_dns.so.2 \
            -l /lib/aarch64-linux-gnu/libresolv.so.2 \
            -l /lib/aarch64-linux-gnu/libgcc_s.so.1; \
    else \
        staticx --strip dist/jinja2_exec dist/jinja2_exec_static \
            -l /lib/x86_64-linux-gnu/libnss_dns.so.2 \
            -l /lib/x86_64-linux-gnu/libresolv.so.2 \
            -l /lib/x86_64-linux-gnu/libgcc_s.so.1; \
    fi