set -e

INSTALL_STARSHIP=${1:-"true"}
USERNAME=${2:-"automatic"}
USER_UID=${3:-"automatic"}
USER_GID=${4:-"automatic"}
INSTALL_PROTOC=${5:-"false"}
INSTALL_KTUNNEL=${6:-"false"}
BASH_HIST_PATH=${7:-"/home/${USERNAME}/.bashhistory"}
MARKER_FILE="/usr/local/etc/vscode-dev-containers/customization"

# Load markers to see which steps have already run
if [ -f "${MARKER_FILE}" ]; then
    echo "Marker file found:"
    cat "${MARKER_FILE}"
    source "${MARKER_FILE}"
fi

# Ensure apt is in non-interactive to avoid prompts
export DEBIAN_FRONTEND=noninteractive

# Run install apt-utils to avoid debconf warning then verify presence of other common developer tools and dependencies
if [ "${PACKAGES_ALREADY_INSTALLED}" != "true" ]; then

    package_list="vim clang-format"

    echo "Packages to verify are installed: ${package_list}"
    apt-get -y install --no-install-recommends ${package_list} 2> >( grep -v 'debconf: delaying package configuration, since apt-utils is not installed' >&2 )

    PACKAGES_ALREADY_INSTALLED="true"
fi

# ** Shell customization section **
if [ "${USERNAME}" = "root" ]; then
    user_rc_path="/root"
else
    user_rc_path="/home/${USERNAME}"
fi

if [ "${PACK_ALREADY_INSTALLED}" != "true" ]; then

    if [ "${INSTALL_STARSHIP}" = "true" ]; then
        curl -Lo "${user_rc_path}/install.sh" "https://starship.rs/install.sh"
        chmod +x ${user_rc_path}/install.sh
        ${user_rc_path}/install.sh -y
        rm ${user_rc_path}/install.sh
    fi

    if [ "${INSTALL_PROTOC}" = "true" ]; then
        protoc_version=$(curl -L -s -H 'Accept: application/json' https://github.com/protocolbuffers/protobuf/releases/latest | sed -e 's/.*"tag_name":"v\([^"]*\)".*/\1/') \
            && DL_URL="https://github.com/protocolbuffers/protobuf/releases/download/v${protoc_version}/protoc-${protoc_version}-linux-x86_64.zip" \
            && curl -LO ${DL_URL} \
            && unzip protoc-${protoc_version}-linux-x86_64.zip -d /home/${USERNAME}/.local \
            && rm *.zip
    fi

    if [ "${INSTALL_KTUNNEL}" = "true" ]; then
        ktunnel_version=$(curl -L -s -H 'Accept: application/json' https://github.com/omrikiei/ktunnel/releases/latest | sed -e 's/.*"tag_name":"v\([^"]*\)".*/\1/') && \
        wget -q https://github.com/omrikiei/ktunnel/releases/download/v${ktunnel_version}/ktunnel_${ktunnel_version}_Linux_x86_64.tar.gz && \
        tar -xf ktunnel_${ktunnel_version}_Linux_x86_64.tar.gz --directory /home/${USERNAME}/.local && \
        rm ktunnel_${ktunnel_version}_Linux_x86_64.tar.gz
    fi

    git clone https://github.com/quickstar/pack.git ${user_rc_path}/.vim/pack --recurse-submodules \
        && ${user_rc_path}/.vim/pack/install.sh "${user_rc_path}"

		## echo "export PATH=${PATH}:/home/${USERNAME}/.dotnet/tools" >> ${user_rc_path}/.bashrc

    chown -R ${USER_UID}:${USER_GID} ${user_rc_path}

    PACK_ALREADY_INSTALLED="true"
fi

# ## Persist bash history between runs
# ## You can also use a mount to persist your command history across sessions / container rebuilds
if [ "${BASH_HIST_PATH}" != "" ]; then
    mkdir ${BASH_HIST_PATH} \
    && touch ${BASH_HIST_PATH}/.bash_history \
    && chown -R $USER_UID:$USER_GID ${BASH_HIST_PATH} \
    && echo "export HISTFILE=${BASH_HIST_PATH}/.bash_history" >> ${user_rc_path}/.bashrc
fi

mkdir -p /home/${USERNAME}/.vscode-server/extensions \
    /home/${USERNAME}/.vscode-server-insiders/extensions

chown -R ${USERNAME}:${USERNAME} /home/${USERNAME}/.vscode-server \
    && chown -R ${USERNAME}:${USERNAME} /home/${USERNAME}/.vscode-server-insiders

# Write marker file
mkdir -p "$(dirname "${MARKER_FILE}")"
echo -e "\
    PACKAGES_ALREADY_INSTALLED=${PACKAGES_ALREADY_INSTALLED}\n\
    PACK_ALREADY_INSTALLED=${PACK_ALREADY_INSTALLED}" > "${MARKER_FILE}"

echo "Done!"
