#!/usr/bin/python

import requests
from ansible.callbacks import vvv

class Connection(object):

    def __init__(self, runner, host, port, user, *args, **kwargs):
        self.runner = runner
        self.host = host
        self.port = port or 8700
        self.user = user
        self.has_pipelining = False

    def _build_url(self, url):
        return 'http://{host}:{port}{url}'.format(host=self.host, port=self.port, url=url)

    def connect(self):
        vvv("ESTABLISH CONNECTION FOR USER: %s" % self.user, host=self.host)
        return self

    def exec_command(self, cmd, tmp_path, *args, **kwargs):
        vvv("EXEC %s" % cmd, host=self.host)

        r = requests.post(self._build_url('/exec'), data={'command': cmd})
        if r.status_code == 200:
            data = r.json()
            return (data['status'], data['stdin'], data['stdout'], data['stderr'])

        return (255, '', '', r.text)

    def put_file(self, in_path, out_path):
        vvv("PUT %s TO %s" % (in_path, out_path), host=self.host)
        with open(in_path, 'rb') as fp:
            r = requests.put(self._build_url('/upload'), data={'dest': out_path}, files={'src': fp})

        if r.status_code != 200:
            raise errors.AnsibleError("failed to transfer file from %s" % in_path)

    def fetch_file(self, in_path, out_path):
        vvv("FETCH %s TO %s" % (in_path, out_path), host=self.host)

    def close(self):
        pass
