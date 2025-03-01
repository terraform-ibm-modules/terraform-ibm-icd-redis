# Restore from backup example

This example provides an end-to-end executable flow of how a Redis DB instance can be created from a backup instance. This example uses the IBM Cloud terraform provider to:

- Create a new resource group if one is not passed in.
- Create a restored ICD Redis database instance pointing to the lastest backup of the existing Redis database instance crn passed.
