#!/bin/bash

#---------------------------------------------------------------
# git-setup.sh
#
# configures git to allow go get to pull from private bitbucket
# repos
#---------------------------------------------------------------

ID_RSA="${HOME}/.ssh/id_rsa"

set -eu

if [ -z "${GIT_SSH_KEY:=}" ] ; then
  exit 0
fi

# configure ssh
#
mkdir -p $(dirname ${ID_RSA})
if [ -f ${ID_RSA} ] ; then
  echo "ssh key already set, ${ID_RSA}"
else
  echo "installing ssh key, ${ID_RSA}"
  cat << EOF > ${ID_RSA}
$(echo ${GIT_SSH_KEY} | base64 --decode)
EOF
  chmod 400 ${ID_RSA}
fi

# automatically accept bitbucket.org ssh keys
#
if grep -q -s "Host bitbucket.org" ${HOME}/.ssh/config ; then
  echo "ssh already configured with Host bitbucket.org"
else
  echo "adding Host bitbucket.org to ${HOME}/.ssh/config"
  cat <<EOF >> ${HOME}/.ssh/config

Host bitbucket.org
  HostName bitbucket.org
  User git
  StrictHostKeyChecking no


EOF
fi

# add git configuration if not present
#
if grep -q -s 'git@bitbucket.org:' ${HOME}/.gitconfig ; then
  echo "git instead of, git@bitbucket.org:, already configured"
else
  echo "configuring git@bitbucket.org: instead of https://bitbucket.org/"
  cat <<EOF >> ${HOME}/.gitconfig
[url "git@bitbucket.org:"]
    insteadOf = https://bitbucket.org/
EOF
fi