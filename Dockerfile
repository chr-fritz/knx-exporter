# Copyright © 2020-2023 Christian Fritz <mail@chr-fritz.de>
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

FROM scratch
COPY scripts/docker/etc_passwd /etc/passwd
COPY knx-exporter /
COPY pkg/.knx-exporter.yaml /etc/knx-exporter.yaml
EXPOSE 8080/tcp
EXPOSE 3671/udp
VOLUME /etc/knx-exporter
USER nonroot
ENTRYPOINT ["/knx-exporter"]
CMD ["run", "--config","/etc/knx-exporter.yaml"]
