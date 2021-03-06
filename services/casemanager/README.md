##Case Manager
The 'casemanager' provides valid testcases with Restful [APIs](API.md) for the users/testserver.

##Case Repos
The official case repo is in [oct](https://github.com/huawei-openlab/oct-engine/cases).
But the casemanager admin can add more repos to the configuration file [casemanager.config](casemanager.config).
Please refer to [How to add a repo](#how-to-add-a-repo] and [How to organize a repo](#how-to-organize-a-repo) to manager your own case repo.

##Cases
The cases in the enabled repo will be validated and providered.
Please refer to [How to add write a testcase](#how-to-write-a-testcase].


- [How to use the `casemanager`](#howto)
- [How does the test case organized](#how-do-the-case-files-been-organized)
- [How to submit a test case](#submit)
- [How to generate the report](reportGenerator/README.md)
- [How to verify the case](caseValidator/HowTO.md)

###HowTo
####HowToRun
Compile the casemanager by using 'Make', modify the [Config](casemanager.conf) to your own preference and run the casemanager.

```
make
./casemanager
```

If you may see the output like this:
```
Error in loading case:  /tmp/casemanager_cache/oct/cases/specstest/namespace  . Skip it
Error in loading case:  /tmp/casemanager_cache/oct/cases/specstest/root_readonly_false  . Skip it
Listen to port  :8011
```
It means there is something wrong with these two cases, please use the [Case Validator Tool](../tools/casevalidator/HowTO.md) to correct them.

####HowToUse
Read the [APIs](API.md)

```
curl localhost:8011/case
```
You may see the output like this:
```
[{"ID":"2f343798066059c41a08ed78d0173d26","Name":"linux_rootfsPropagation","GroupDir":"/tmp/casemanager_cache/oct/cases/specstest","LibFolderName":"source","Status":"idle","TestedTime":0,"LastModifiedTime":1439693492},{"ID":"b9610144452ea7fa91cc17e766e3df8d","Name":"mount_fstype","GroupDir":"/tmp/casemanager_cache/oct/cases/specstest","LibFolderName":"source","Status":"idle","TestedTime":0,"LastModifiedTime":1439600364},{"ID":"ae4984cb1287618c52d3a38ee0ac8c23","Name":"platform_linux_amd64","GroupDir":"/tmp/casemanager_cache/oct/cases/specstest","LibFolderName":"source","Status":"idle","TestedTime":0,"LastModifiedTime":1439602673},{"ID":"e5ce923b725d841c179646b94b2dabde","Name":"root_readonly_true","GroupDir":"/tmp/casemanager_cache/oct/cases/specstest","LibFolderName":"source","Status":"idle","TestedTime":0,"LastModifiedTime":1439693492},{"ID":"ef3101d50a06111a91529b7c9264310c","Name":"version_error","GroupDir":"/tmp/casemanager_cache/oct/cases/specstest","LibFolderName":"source","Status":"idle","TestedTime":0,"LastModifiedTime":1439602544},{"ID":"6014fa2a802dad3cc69c43fc410ec790","Name":"rktcpumonitor","GroupDir":"/tmp/casemanager_cache/oct/cases/benchmark/monitor","LibFolderName":"source","Status":"idle","TestedTime":0,"LastModifiedTime":1439600364},{"ID":"004dbf6f5b627c0e60c90559d3c97322","Name":"linux_capabilites","GroupDir":"/tmp/casemanager_cache/oct/cases/specstest","LibFolderName":"source","Status":"idle","TestedTime":0,"LastModifiedTime":1439693492},{"ID":"d9ad9e2a4d840610eaadb8226c169ae7","Name":"process","GroupDir":"/tmp/casemanager_cache/oct/cases/specstest","LibFolderName":"source","Status":"idle","TestedTime":0,"LastModifiedTime":1439693492},{"ID":"5e1bde28cdf2ea66b41c7a7508f744cf","Name":"version_correct","GroupDir":"/tmp/casemanager_cache/oct/cases/specstest","LibFolderName":"source","Status":"idle","TestedTime":0,"LastModifiedTime":1439602585}]
```

If you want to get the files of a single case, you can:

```
curl localhost:8011/case/5e1bde28cdf2ea66b41c7a7508f744cf > version_correct.tar.gz
```

###How do the case files been organized
The ideal case struct should be like this [case bundle](https://github.com/huawei-openlab/oct-engine/blob/master/cases/spec.md#case-bundle):

```
casedir
  |__ group one
  |        |_____ case one
  |        |_____ case two
  |
  |__ group two
  |__ group three
	   |____  case three
			|___ case.json
			|___ source/container-file
			|___ source/container-script one
			|___ report.md

```

But currently, the following flexiable format is also supported. (will be remove soon)

```
casedir
  |__ group one
  |           |___ case four
  |                      |____  case-four-name.json
  |                      |____  source/case-four-files
  |___ group two
              |___  case five
              |           |____ case-five-name.json
              |           |____ source/case-five-files
              |___  case six
              |___  LibOne  (LibOne is used in case-five and case-six
                          |____ files
                          
```

In the `flexiable format`, we need to configurate the `Test Case Server`, for example:
[Demo configuration](casemanager.conf)
```

{
	"GitRepo":  "https://github.com/huawei-openlab/oct.git",
        "CaseFolderName":  "cases",
	"Groups": [ 
		{"Name": "specstest",
		 "LibFolderName": "source"
		},
		{"Name": "benchmark/monitor",
		 "LibFolderName" : "source"
		}
	],
	"CacheDir": "/tmp/casemanager_cache/",
	"Port":  8011
}
```

###submit
Before submit case to the repo, please use the [Case Validator Tool](../tools/casevalidator/HowTO.md) to check its validation.

After that, please submit to this OCT repo and put your case under the 'cases' directory.
