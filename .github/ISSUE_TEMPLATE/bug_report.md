---
name: Bug report
about: Create a report to help us improve
title: ''
labels: bug
assignees: ''

---

### Terraform Version
Run `terraform -v` to show the version. If you are not running the latest version of Terraform, please upgrade because your issue may have already been fixed.

### Affected Resource(s)
Please list the resources as a list, for example:
- hopsworksai_cluster

If this issue appears to affect multiple resources, it may be an issue with Terraform's core, so please mention this.

### Terraform Configuration Files
```hcl
# Copy-paste your Terraform configurations here 
```

### Debug Output
Please provider a link to a GitHub Gist containing the complete debug output: https://www.terraform.io/docs/internals/debugging.html. Please do NOT paste the debug output in the issue; just paste a link to the Gist.

### Panic Output
If Terraform produced a panic, please provide a link to a GitHub Gist containing the output of the `crash.log`.

### Expected Behavior
What should have happened?

### Actual Behavior
What actually happened?

### Steps to Reproduce
Please list the steps required to reproduce the issue, for example:
1. `terraform apply`

### Important Factoids
Are there anything atypical about your accounts that we should know? 

### References
Are there any other GitHub issues (open or closed) or Pull Requests that should be linked here? For example:
- GH-1234
