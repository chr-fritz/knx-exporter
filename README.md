[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fchr-fritz%2Fknx-exporter.svg?type=small)](https://app.fossa.com/projects/git%2Bgithub.com%2Fchr-fritz%2Fknx-exporter?ref=badge_small)

# KNX Prometheus Exporter

The KNX Prometheus Exporter is a small bridge to export values measured
by KNX sensors to Prometheus. It takes the values either from cyclic
sent `GroupValueWrite` telegrams and can request values itself using
`GroupValueRead` telegrams.

[TOC]: # "## Table of Contents"

## Table of Contents
- [KNX Prometheus Exporter](#knx-prometheus-exporter)
  - [Table of Contents](#table-of-contents)
  - [Usage](#usage)
    - [Preparing the configuration](#preparing-the-configuration)
  - [Contributing](#contributing)
  - [License](#license)
  - [Maintainer](#maintainer)


## Usage

### Preparing the configuration

The KNX Prometheus Exporter will only export the values from configured
group addresses. A good starting point is to
[export the group addresses](https://support.knx.org/hc/en-us/articles/115001825324-Group-Address-Export)
from ETS 5 into the XML format and convert them.

Please refer the KNX Documentation
"[Group Address & Export](https://support.knx.org/hc/en-us/articles/115001825324-Group-Address-Export)"
for more information about how to export the group addresses to XML.

With the exported group addresses you can call the `knx-exporter` and
convert them:

```shell script
knx-exporter convertGA [SOURCE] [TARGET]
```

You must replace `[SOURCE]` with the path to your group address export
file. `[TARGET]` is the path where the converted configuration should be
stored.


## Contributing

1. Fork it
2. Download your fork to your PC (`git clone
   https://github.com/your_username/knx-exporter && cd knx-exporter`)
3. Create your feature branch (`git checkout -b my-new-feature`)
4. Make changes and add them (`git add .`)
5. Commit your changes (`git commit -m 'Add some feature'`)
6. Push to the branch (`git push origin my-new-feature`)
7. Create new pull request

## License

The KNX Exporter is released under the Apache 2.0 license. See
[LICENSE](https://github.com/chr-fritz/knx-exporter/blob/master/LICENSE)

## Maintainer

Christian Fritz (@chrfritz)
