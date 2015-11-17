#!/usr/bin/python

import requests
from ansible import utils
from ansible.callbacks import vvv
from ansible.constants import p, get_config

DEFAULT_USE_SSL = get_config(p, 'agent', 'use_ssl', None, False, boolean=True)

class Connection(object):

    def __init__(self, runner, host, port, user, *args, **kwargs):
        self.runner = runner
        self.host = host
        self.port = port or 8700
        self.user = user
        self.proto = 'http'
        if DEFAULT_USE_SSL:
            self.proto = 'https'
        self.has_pipelining = False

    def _build_url(self, url):
        return '{proto}://{host}:{port}{url}'.format(proto=self.proto, host=self.host, port=self.port, url=url)

    def connect(self):
        vvv("ESTABLISH CONNECTION FOR USER: %s" % self.user, host=self.host)
        self.session = requests.Session()
        self.session.verify = False
        return self

    def exec_command(self, cmd, tmp_path, *args, **kwargs):
        vvv("EXEC %s" % cmd, host=self.host)

        data = {'command': cmd}
        executable = kwargs.get('executable')
        if executable is not None:
            data['executable'] = executable

        if self.runner.become:
            data['become'] = 1
            if self.runner.become_method:
                data['becomeMethod'] = self.runner.become_method

        r = self.session.post(self._build_url('/exec'), data=data)
        if r.status_code == 200:
            data = r.json()
            return (data['status'], data['stdin'], data['stdout'], data['stderr'])

        return (255, '', '', r.text)

    def put_file(self, in_path, out_path):
        vvv("PUT %s TO %s" % (in_path, out_path), host=self.host)
        with open(in_path, 'rb') as fp:
            r = self.session.put(self._build_url('/upload'), data={'dest': out_path}, files={'src': fp})

        if r.status_code != 200:
            raise errors.AnsibleError("failed to transfer file from %s" % in_path)

    def fetch_file(self, in_path, out_path):
        vvv("FETCH %s TO %s" % (in_path, out_path), host=self.host)
        raise errors.AnsibleError("not unimplemented")

    def close(self):
        self.session.close()
