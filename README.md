# aws-one-punch
One punch to grant all command line windows AWS access in MacOS

## Background ##
AWS provides three options to access the resources:
* management console access
* command line access
* programmatic access

When working with micro services using AWS resources like Cloudformation, we normally open multiple command line windows so that we can update the corresponding AWS resources after changing each individual service. The pain point is we have to grant the access every one hour by setting the AWS environment variables in each command line window. One solution is to update the local credentials file instead of setting the environment variables, but that has to be done manually, every one hour.

## Solution ##
AWS-one-punch basically pulls all accounts and profiles with the token stored in the cookie, and generate new credentials and get them updated in the local credentials file, thus with just one command, we can grant all command line windows AWS access.

## Setup ##
1. install with Homebrew
 ```
   brew tap gaarazhu/aws-one-punch && brew install gaarazhu/aws-one-punch/aws-one-punch
 ```
2. set AWS management console domain in `~/.bash_profile` and reload it with `source ~/.bash_profile`
```
   export AWS_CONSOLE_DOMAIN="garyz.awsapps.com"
 ```
3. run
 ```
$ aws-one-punch
NAME:
   aws-one-punch - one punch to access AWS resoruces in command line

USAGE:
   aws-one-punch [global options] command [command options] [arguments...]

VERSION:
   1.0.0

COMMANDS:
   list-accounts, ls-a  List accounts
   list-profiles, ls-p  List profiles under an account
   access, access       Access AWS Resource with a profile
   help, h              Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
```

## Usage ##
1. list accounts
```
$ aws-one-punch list-accounts
2021/11/10 22:04:14 no AWS SSO Token found, please open the AWS Management Console https://gzhu.awsapps.com/start/#/ first
```

2. open the url, wait for the SSO finished and run above command again(PS. keep listing the accounts unitl it works as the token will only be written to local cookie after all resources have been loaded during the SSO)
```
$ aws-one-punch list-accounts
AccountId: ins-sd4312, accountName: 20890663 (MRP IaaS Prod)
AccountId: ins-2ssfds, accountName: 79300001 (Sandbox 1)
AccountId: ins-3sadfa, accountName: 69127290 (MRP IaaS Non-Prod)
AccountId: ins-siki23, accountName: 58868209 (Data Analytics)
AccountId: ins-14oasn, accountName: 66060440 (Shared Services)
```

3. list profiles
```
$ aws-one-punch list-profiles ins-3sadfa
ProfileName: DigitalDeveloperNonprodAccess
```
4. one punch for access
```
$ aws-one-punch access 69127290 DigitalDeveloperNonprodAccess
AWS access granted with account 69127290 and profile DigitalDeveloperNonprodAccess
```

## Contribution ##
Your contributions are always welcome!

## License ##
This work is licensed under [Creative Commons Attribution 4.0 International Licenese](https://creativecommons.org/licenses/by/4.0/).
