ARG GOLANG_VERSION=1.18
FROM golang:${GOLANG_VERSION}-bullseye

# [Option] Install zsh
ARG INSTALL_ZSH="false"
# [Option] Install oh-my-bash or zsh
ARG INSTALL_OH_MYS="false"
# [Option] Upgrade OS packages to their latest versions
ARG UPGRADE_PACKAGES="true"
# [Option] Enable non-root Docker access in container
ARG ENABLE_NONROOT_DOCKER="true"
# [Option] Needed for adding manpages-posix and manpages-posix-dev which are non-free packages in Debian
ARG ADD_NON_FREE_PACKAGES="true"

# Install needed packages and setup non-root user. Use a separate RUN statement to add your own dependencies.
ARG SOURCE_SOCKET=/var/run/docker-host.sock
ARG TARGET_SOCKET=/var/run/docker.sock

ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=$USER_UID

COPY library-scripts/*.sh /tmp/library-scripts/

RUN apt-get update
RUN export DEBIAN_FRONTEND=noninteractive \
	&& bash /tmp/library-scripts/common-debian.sh "${INSTALL_ZSH}" "${USERNAME}" "${USER_UID}" "${USER_GID}" "${UPGRADE_PACKAGES}" \
	&& apt-get -y install --no-install-recommends lynx
# Use Docker script from script library to set things up
RUN export DEBIAN_FRONTEND=noninteractive \
	&& /bin/bash /tmp/library-scripts/docker-debian.sh "${ENABLE_NONROOT_DOCKER}" "${SOURCE_SOCKET}" "${TARGET_SOCKET}" "${USERNAME}"

# Install GoLang and related tools
ARG GOPATH="/go"
RUN export DEBIAN_FRONTEND=noninteractive && \
	if [ "$GOLANG_VERSION" != "" ]; then /bin/bash /tmp/library-scripts/go-debian.sh \
	"${GOLANG_VERSION}" \
	"/usr/local/go" \
	"${GOPATH}" \
	"${USERNAME}"; \
	fi

# [Option] Install project specific customizations
ARG INSTALL_CUSTOMIZATION="true"
ARG INSTALL_STARSHIP="true"
ARG INSTALL_PROTOC="false"
ARG INSTALL_KTUNNEL="false"
RUN export DEBIAN_FRONTEND=noninteractive && \
	if [ "$INSTALL_CUSTOMIZATION" = "true" ]; then /bin/bash /tmp/library-scripts/customization-debian.sh \
	"${INSTALL_STARSHIP}" \
	"${USERNAME}" \
	"${USER_UID}" \
	"${USER_GID}"; \
	fi

# Clean up
RUN apt-get autoremove -y && apt-get clean -y && rm -rf /var/lib/apt/lists/* /tmp/library-scripts/ /tmp/config/

# Setting the ENTRYPOINT to docker-init.sh will configure non-root access to
# the Docker socket if "overrideCommand": false is set in devcontainer.json.
# The script will also execute CMD if you need to alter startup behaviors.
ENTRYPOINT [ "/usr/local/share/docker-init.sh" ]
CMD [ "sleep", "infinity" ]
