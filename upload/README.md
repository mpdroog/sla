Upload
==============
Upload directory to Usenet and output NZB.

* Read files from `config.Uploadir`
* ZIP them
* yEnc encode them
* Upload to `config.Address` 
* Output NZB for future downloading in `config.NzbDir`

> This tools required at least 50 articles to work.
> 50 articles * 750KB +- 36MB

config.json
```
{
	"Address": "127.0.0.1:9091",
	"User": "user",
	"Pass": "test",
	"NzbDir": "./",
	"MsgDomain": "@usenet.farm",
	"UploadDir": "./dummy"
}
```
Dummy(mock) server available on https://github.com/mpdroog/spool-mock

