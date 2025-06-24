FROM ubuntu:20.04

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

# 安装PyInstaller和StaticX
RUN pip3 install --upgrade pip
RUN pip3 install setuptools wheel
RUN pip3 install pyinstaller
RUN pip3 install staticx

# 安装impacket
RUN pip3 install impacket

COPY wmiexec.py /app/

# 使用PyInstaller构建可执行文件
RUN pyinstaller --onefile --clean --strip --noupx wmiexec.py

# 提供ARG用于检测当前架构
ARG TARGETARCH

# 根据不同架构使用不同的库路径
RUN if [ "$TARGETARCH" = "amd64" ]; then \
        staticx --strip dist/wmiexec dist/wmiexec-static \
            -l /lib/x86_64-linux-gnu/libnss_dns.so.2 \
            -l /lib/x86_64-linux-gnu/libresolv.so.2 \
            -l /lib/x86_64-linux-gnu/libgcc_s.so.1; \
    elif [ "$TARGETARCH" = "arm64" ]; then \
        staticx --strip dist/wmiexec dist/wmiexec-static \
            -l /lib/aarch64-linux-gnu/libnss_dns.so.2 \
            -l /lib/aarch64-linux-gnu/libresolv.so.2 \
            -l /lib/aarch64-linux-gnu/libgcc_s.so.1; \
    elif [ "$TARGETARCH" = "arm" ]; then \
        staticx --strip dist/wmiexec dist/wmiexec-static \
            -l /lib/arm-linux-gnueabihf/libnss_dns.so.2 \
            -l /lib/arm-linux-gnueabihf/libresolv.so.2 \
            -l /lib/arm-linux-gnueabihf/libgcc_s.so.1; \
    else \
        staticx --strip dist/wmiexec dist/wmiexec-static \
            -l /lib/x86_64-linux-gnu/libnss_dns.so.2 \
            -l /lib/x86_64-linux-gnu/libresolv.so.2 \
            -l /lib/x86_64-linux-gnu/libgcc_s.so.1; \
    fi