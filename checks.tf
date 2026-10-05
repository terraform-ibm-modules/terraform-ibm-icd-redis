##############################################################################
# Check Blocks
##############################################################################

check "warn_hs_crypto_key" {
  assert {
    condition = !(
      (var.kms_key_crn != null && can(regex(".*hs-crypto.*", var.kms_key_crn))) ||
      (var.backup_encryption_key_crn != null && can(regex(".*hs-crypto.*", var.backup_encryption_key_crn)))
    )
    error_message = "WARNING: An IBM Cloud Hyper Protect Crypto Services (hs-crypto) key CRN was provided. Note that IBM Cloud Hyper Protect Crypto Services is set to be deprecated soon."
  }
}
