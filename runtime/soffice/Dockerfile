FROM debian:11-slim

WORKDIR /var/tmp/

# Install dependencies
RUN apt-get update && apt-get install -y libxinerama1 libnss3 libdbus-1-3 libgio-cil libcairo2 libcups2 openjdk-17-jre-headless && rm -rf /var/lib/apt/lists/*

# Install LibreOffice
ADD https://download.documentfoundation.org/libreoffice/stable/7.4.0/deb/x86_64/LibreOffice_7.4.0_Linux_x86-64_deb.tar.gz .
ADD https://download.documentfoundation.org/libreoffice/stable/7.4.0/deb/x86_64/LibreOffice_7.4.0_Linux_x86-64_deb_sdk.tar.gz .
RUN tar -xf LibreOffice_7.4.0_Linux_x86-64_deb.tar.gz && \
    tar -xf LibreOffice_7.4.0_Linux_x86-64_deb_sdk.tar.gz && \
    dpkg -i LibreOffice_7.4.0.3_Linux_x86-64_deb/DEBS/*.deb && \
    dpkg -i LibreOffice_7.4.0.3_Linux_x86-64_deb_sdk/DEBS/*.deb && \
    rm -rf /var/tmp/*

ENV PATH=/opt/libreoffice7.4/program/:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
ADD entrypoint.sh /usr/local/bin/

# We will propably also use this image for having the SDK there,
# so we make this entrypoint script only a convenient CMD.
CMD ["/usr/local/bin/entrypoint.sh"]
