# Netreduce

Netreduce is an API aggregator for HTTP services. Its primary goal is to provide an adapter layer between
multiple different backend services and their clients. Netreduce provides an interface that can be optimized for
the requirements of the clients, while allowing a clean and normalized interface on the backend services that
are the owners of the actual resources. One typical use case of netreduce is the BFF scenario (Backend For the
Frontend), where it can be used as an alternative to creating a custom adapter service from scratch.

### Features

** WIP: netreduce is a work-in-progress project, the below features are meant as currently planned and can be in
different state of availability, can be changed, and finally also can be dropped, until the first beta version
of netreduce is released. **

- many-to-many releation between backend services and frontend endpoints
- automatic parallelization/optimization of backend requests
- safe definition of frontend endpoints without downtime
- safe extensibility with custom backend connectors
- safe extensibility with custom mapping functions for the frontend endpoints
