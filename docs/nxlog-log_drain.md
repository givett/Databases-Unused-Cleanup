##Adding nxlog into RDPG

For  CI/Pipeline based deployments:

1.  DL the current 0.1.0 release tarball of nxlog : https://github.com/hybris/nxlog-boshrelease/releases/download/v0.1.0/nxlog-0.1.0.tgz
2.  Target your BOSH Director
3.  Upload the nxlog release ```bosh upload release nxlog-0.1.0.tgz```
4.  Modify your deployment templates to include the release, jobs, and the endpoint settings. For CI deployments:
      * Release changes - stub,yml
        * Add under releases:
         ```
           - name: nxlog
             version: latest
        ```
        * Template changes - jobs.yml
          * Add under templates for each deployment job:
          ```
          - name: nxlog
            release: nxlog
          ```
        * Manifest changes - monitoring.yml
          * Add the environment specific connection info:
          ```
          nxlog:
            tcpoutput:
              host: 127.0.0.1
              port: 8080
          ```

These changes are for CI/Pipeline based deployments

For Bosh-Lite integration:

     Make your changes in /templates/jobs.yml and stub.yml

All nxlog logs are written out under /var/vcap/sys/log/nxlog - there should be 3 total.
The main ends in .log, the other 2 are stdout and stderror.
Main one to watch is .log and make sure it's connecting to the log service.
Give it a good 10 min or so before you check the log search site for the logs to be parsed and ready for searching/viewing.
