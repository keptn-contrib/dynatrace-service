## Known Limitations

* The Dynatrace Metrics API provides data with the "eventually consistency" approach. Therefore, the metrics data retrieved can be incomplete or even contain inconsistencies in case of time frames that are within two hours of the current datetime. Usually, it takes a minute to catch up, but in extreme situations this might not be enough. We try to mitigate that by delaying calls to the metrics API by 60 seconds.

* This service uses the Dynatrace Metrics v2 API by default but can also parse v1 metrics query. If you use the v1 query language you will see warning log outputs in the *dynatrace-service* which encourages you to update your queries to v2. More information about Metrics v2 API can be found in the [Dynatrace documentation](https://www.dynatrace.com/support/help/extend-dynatrace/dynatrace-api/environment-api/metric-v2/)
