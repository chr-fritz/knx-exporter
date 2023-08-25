#!/bin/sh
#
# Copyright © 2022-2023 Christian Fritz <mail@chr-fritz.de>
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

cleanInstall() {
    printf "\033[32m Post Install of an clean install\033[0m\n"
    # Step 3 (clean install), enable the service in the proper way for this platform
    echo "# Remove this file to allow auto starts of the knx-exporter daemon through systemd.\n" >/etc/knx-exporter/knx-exporter_not_to_be_run

    printf "\033[32m Reload the service unit from disk\033[0m\n"
    systemctl daemon-reload || :
    printf "\033[32m Unmask the service\033[0m\n"
    systemctl unmask knx-exporter.service || :
    printf "\033[32m Set the preset flag for the service unit\033[0m\n"
    systemctl preset knx-exporter.service || :
    printf "\033[32m Set the enabled flag for the service unit\033[0m\n"
    systemctl enable knx-exporter.service || :
    systemctl restart knx-exporter.service || :
}

upgrade() {
    printf "\033[32m Post Install of an upgrade\033[0m\n"
    systemctl daemon-reload || :
    systemctl restart knx-exporter.service || :
}

# Step 2, check if this is a clean install or an upgrade
action="$1"
if [ "$1" = "configure" ] && [ -z "$2" ]; then
    # Alpine linux does not pass args, and deb passes $1=configure
    action="install"
elif [ "$1" = "configure" ] && [ -n "$2" ]; then
    # deb passes $1=configure $2=<current version>
    action="upgrade"
fi

case "$action" in
"1" | "install")
    cleanInstall
    ;;
"2" | "upgrade")
    printf "\033[32m Post Install of an upgrade\033[0m\n"
    upgrade
    ;;
*)
    # $1 == version being installed
    printf "\033[32m Alpine\033[0m"
    cleanInstall
    ;;
esac
