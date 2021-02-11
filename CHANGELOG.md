# CHANGELOG

## 2.0.3 (February 11, 2021)

- Add puppetca_certificate resource import functionality
- Use go-puppetca/GetCertStatusByName to populate certificate metadata properly in state

## 2.0.2 (February 10, 2021)

- Update go-puppetca dependency
- Update terraform-plugin-sdk to v2

## 2.0.1 (July 9, 2020)

BREAKING CHANGES:

- Return structured certificate info rather than raw PEM.
- Use forked go-puppetca module to support enchancements.

ENHANCEMENTS:

- Enable automatic certificate signing during resource creation.
- Add ignore_ssl provider parameter.

## 1.3.0 (July 8, 2020)

- Add ignore_ssl provider parameter
- Add certificate signing logic

## 1.2.2 (June 05, 2019)

- Use Terraform v0.12's API

## 1.2.1 (May 20, 2019)

- Fix build config

## 1.2.0 (May 20, 2019)

- Use Go modules
- Use Terraform v0.12-beta's API

## 1.1.0 (Oct 17, 2018)

- rebuild with new go-puppetca library (support for SSL parameters as strings)

## 1.0.1 (Oct 13, 2017)

IMPROVEMENTS:

- resource/puppetca_certificate: increase NotFoundChecks ([#1](https://github.com/greennosedmule/terraform-provider-puppetca/issues/1))

## 1.0.0 (Sept 29, 2017)

- Initial release
