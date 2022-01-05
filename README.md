# aws-one-punch
One punch to grant all command prompts AWS access with IAM role credentials in OSX.

## Background ##
When working with micro services using Cloudformation, we normally open multiple command prompts to update the corresponding AWS resources after changing each individual service. As recommended by AWS, we should use IAM roles instead of long-term access keys in this case. But the pain point is that we will have to grant the access in each command prompt, or to update the local credentials file every time when the temporary credentials are expired.

## Solution ##
AWS-one-punch pulls all AWS accounts and IAM roles with the SSO bearer token stored in cookie to generate new credentials to be updated in the local credentials file, thus with just one command, we can grant all command prompts the access.

## Prerequisites ##
AWS CLI needs to be installed and configured.

**Note: for simplicity, the `default` profile will be used for one punch access.**

## Setup ##
1. install via Homebrew
 ```
   brew install gaarazhu/aws-one-punch/aws-one-punch
 ```
2. set the AWS user portal in `~/.bash_profile` or equivalent and reload it with `source ~/.bash_profile`
```
   export AWS_CONSOLE_DOMAIN="garyz.awsapps.com"
 ```
3. run
 ```
$ aws-one-punch
NAME:
   aws-one-punch - one punch to grant all command prompts AWS access with IAM role credentials in OSX.

USAGE:
   aws-one-punch [global options] command [command options] [arguments...]

VERSION:
   1.0.0

COMMANDS:
   list-accounts, ls-a  List accounts
   list-roles, ls-r     List IAM roles under an account
   access, a            Access AWS Resource with IAM role credentials
   help, h              Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
```

## Usage ##
1. list accounts
```
$ aws-one-punch list-accounts
2021/11/10 22:04:14 No AWS SSO token found, please finish the SSO in the user portal first: https://gzhu.awsapps.com/start/#/ first
```

2. open the url, finish SSO and repeat step one(PS. keep listing the accounts unitl it works as the token will only be written to cookie when all resources have been loaded after the SSO)
```
$ aws-one-punch list-accounts
AccountId: ins-sd4312, accountName: 20890663 (MRP IaaS Prod)
AccountId: ins-2ssfds, accountName: 79300001 (Sandbox 1)
AccountId: ins-3sadfa, accountName: 69127290 (MRP IaaS Non-Prod)
AccountId: ins-siki23, accountName: 58868209 (Data Analytics)
AccountId: ins-14oasn, accountName: 66060440 (Shared Services)
```

3. list IAM roles
```
$ aws-one-punch list-roles --account-id ins-3sadfa
RoleName: DigitalDeveloperNonprodAccess
```
4. one punch for access
```
$ aws-one-punch access --account-name 69127290 --role-name DigitalDeveloperNonprodAccess
AWS access granted with account 69127290 and IAM role DigitalDeveloperNonprodAccess
```

## Contribution ##
Your contributions are always welcome!

## License ##
This work is licensed under [MIT](https://opensource.org/licenses/MIT).
