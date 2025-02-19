moved {
  from = module.redis.module.redis.ibm_database.redis_database
  to   = module.redis[0].module.redis.ibm_database.redis_database
}

moved {
  from = module.redis.module.redis.ibm_iam_authorization_policy.kms_policy[0]
  to   = module.redis[0].module.redis.ibm_iam_authorization_policy.kms_policy[0]
}

moved {
  from = module.redis.module.redis.time_sleep.wait_for_authorization_policy[0]
  to   = module.redis[0].module.redis.time_sleep.wait_for_authorization_policy[0]
}

moved {
  from = module.redis.module.redis.ibm_resource_tag.access_tag[0]
  to   = module.redis[0].module.redis.ibm_resource_tag.access_tag[0]
}
