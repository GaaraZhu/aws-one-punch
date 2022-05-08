# aws-one-punch
One command to grant all command prompts AWS access with IAM role credentials in OSX for AWS SSO users. 

## Background ##
When working in organizations where [AWS SSO](https://aws.amazon.com/single-sign-on/) is used to manage the AWS accounts, we can obtain the [IAM temporary credentials](https://docs.aws.amazon.com/singlesignon/latest/userguide/howtogetcredentials.html) in the user portal for command line or programmatic access to the cloud resources. The pain point is the manual work (generate credentials, copy paste and execute in the command prompt, or to update the local credentials file) needs to be done every time when the temporary credentials are expired, and it will become worse when there are multiple command prompts opened which is quite common when working with micro services whose resources are maintained through [CloudFormation](https://aws.amazon.com/cloudformation/) or equivalent.

## Solution ##
AWS-one-punch retrieves the AWS SSO bearer token stored in Chrome cookie after the authentication process to provide below functionalities:
* List all assigned AWS accounts
* List all assigned AWS IAM role in an AWS account
* Grant all command promopts AWS access with temporary credentails from an IAM role

## Prerequisites ##
* AWS CLI needs to be installed and configured.
* Chrome is used for AWS SSO.

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
1. List all assigned AWS accounts
```
$ aws-one-punch list-accounts
AccountId: ins-sd4312, accountName: 20890663 (MRP IaaS Prod)
AccountId: ins-2ssfds, accountName: 79300001 (Sandbox 1)
AccountId: ins-3sadfa, accountName: 69127290 (MRP IaaS Non-Prod)
AccountId: ins-siki23, accountName: 58868209 (Data Analytics)
AccountId: ins-14oasn, accountName: 66060440 (Shared Services)
```

2. List all assigned AWS IAM role in an AWS account
```
$ aws-one-punch list-roles --account-id ins-3sadfa
RoleName: DigitalDeveloperNonprodAccess
```

3. Grant all command promopts AWS access with temporary credentails from an IAM role
```
$ aws-one-punch access --account-name 69127290 --role-name DigitalDeveloperNonprodAccess
AWS access granted with account 69127290 and IAM role DigitalDeveloperNonprodAccess
```

## Simplification ##
For furthur simplification, we can create an [alias](https://wpbeaches.com/make-an-alias-in-bash-or-zsh-shell-in-macos-with-terminal/) for above access command, or have it managed through [pet](https://github.com/knqyf263/pet).

## Limitation ##
There is a delay up to 30 seconds after the SSO authentication before the token is available in the Cookie due to Chrome's persistence implementation with [SQLitePersistentCookieStore](https://www.chromium.org/developers/design-documents/network-stack/cookiemonster/). If the same error message is showing after the SSO authentication, please keep trying until it works.
```
2021/11/10 22:04:14 No AWS SSO token found, please finish the SSO in the user portal first: https://gzhu.awsapps.com/start/#/ first
```

## Contribution ##
Your contributions are always welcome!

## License ##
This work is licensed under [MIT](https://opensource.org/licenses/MIT).
