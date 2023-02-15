# FS Cloud profile example

An example using the fscloud profile to deploy a compliant Redis instance:
 - Create a new resource group if one is not passed in.
 - Create a new redis database instance.
 - Create Key Protect instance with root key.
 - Backend encryption using generated Key Protect key.
 - Create a Sample VPC.
 - Create Context Based Restriction(CBR) to only allow Redis to be accessible from the VPC.
