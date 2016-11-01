FROM ubuntu
MAINTAINER esell <esell@esheavyindustries.com>
RUN apt-get update && apt-get install -y wget vim curl rsyslog build-essential && \
    cd /tmp && wget -q -O snoopy-install.sh https://github.com/a2o/snoopy/raw/install/doc/install/bin/snoopy-install.sh && \
    chmod 755 snoopy-install.sh && ./snoopy-install.sh stable && rm -rf * && rm -rf /var/lib/apt/lists/*
ADD snoopy.ini /etc/
