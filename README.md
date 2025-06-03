### A cli tool to track couriers

#### Usage examples:

- track the courier with `speedaf`, refreshed automatically after 5 seconds:
``` bash
trakr -trackingNumber abc123 -service speedaf -refreshInterval 5
```


- show tracking info, then exits
``` bash
trakr -trackingNumber abc123 -service speedaf
```

### Supported Trackers:

Supported services can be found in [services](https://github.com/ukashazia/trakr/blob/9b04f822bde4463c2758df16be4b454fd5b0fe8c/trackers/trackers.go#L8)
