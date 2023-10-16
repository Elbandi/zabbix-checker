#!/usr/bin/env python3
import argparse
import json
import re
from enum import Enum
from dotenv import dotenv_values
import sys
import os

import proxmoxer


class QueryPBS:
    def __init__(self):
        self._modes = {
            "datastores": self.get_datastores,
            "groups": self.get_groups,
            "snapshots": self.get_snapshots
        }
        default_endpoint = "127.0.0.1"
        default_username = None
        default_tokenname = None
        if "PBS_REPOSITORY" in os.environ:
            repo = os.getenv("PBS_REPOSITORY", "")
            default_endpoint, default_username, default_tokenname = self._parse_env(repo)

        parser = argparse.ArgumentParser("check_pbs")
        parser.add_argument("--env-file", type=str,
                            help="Read env variables from this file")
        parser.add_argument("--api-endpoint", "-e", type=str,
                            default=default_endpoint,
                            help="PBS api endpoint host")
        parser.add_argument("--api-port", type=int, default=8007,
                            help="PBS api endpoint port")
        parser.add_argument("--username", "-u", type=str,
                            default=default_username,
                            help="PBS api user (root@pam, icinga2@pve, ...)")
        parser.add_argument("--password", "-p", type=str,
                            default=os.environ.get('PBS_PASSWORD'),
                            help="PBS api user password")
        parser.add_argument("--token-name", type=str,
                            default=default_tokenname,
                            help="PBS api token name")
        parser.add_argument("--token-value", type=str,
                            default=os.environ.get('PBS_PASSWORD'),
                            help="PBS api token value")
        parser.add_argument("--insecure", "-k", action="store_true",
                            default=False,
                            help="Don't verify HTTPS certificate")
        parser.add_argument("--mode", "-m",
                            choices=self._modes.keys(),
                            help="Check mode to use.")
        parser.add_argument("--exclude", "-E", action='append', default=[],
                            help="Exclude specified resource")
        self._parser = parser
        self._args = parser.parse_args()

        if self._args.env_file:
            denv = dotenv_values(self._args.env_file)
            if "PBS_REPOSITORY" in denv:
                repo = denv.get("PBS_REPOSITORY")
                api_endpoint, username, tokenname = self._parse_env(repo)
                if self._args.api_endpoint is None or self._args.api_endpoint == default_endpoint:
                    self._args.api_endpoint = api_endpoint
                if self._args.username is None:
                    self._args.username = username
                if self._args.token_name is None:
                    self._args.token_name = tokenname
            if "PBS_PASSWORD" in denv:
                self._args.password = denv.get("PBS_PASSWORD")
                self._args.token_value = denv.get("PBS_PASSWORD")

        self._connect()

    def _parse_env(self, repo):
        m = re.search("((?P<user>\\w+@\\w+)(?:!(?P<token>\\w+))?@)?(?P<host>\\S+)", repo)
        if m:
            return m.group('host'), m.group('user'), m.group('token')
        return None, None, None

    def _connect(self):
        if self._args.token_name and self._args.token_value:
            self._pbs = proxmoxer.ProxmoxAPI(self._args.api_endpoint, service="PBS",
                                             user=self._args.username, token_name=self._args.token_name,
                                             token_value=self._args.token_value,
                                             verify_ssl=not self._args.insecure)
        else:
            self._pbs = proxmoxer.ProxmoxAPI(self._args.api_endpoint, service="PBS",
                                             user=self._args.username, password=self._args.password,
                                             verify_ssl=not self._args.insecure)

    def _get_datastore_usage(self):
        return self._pbs.status("datastore-usage").get()

    def _get_group_name(self, store, ns):
        if ns:
            return store + "/" + ns
        return store

    def get_groups(self):
        datastore_usage = self._get_datastore_usage()
        groups = {}
        for ds in sorted(datastore_usage, key=lambda x: x["store"]):
            if ds["store"] in self._args.exclude:
                continue
            namespaces = self._pbs.admin.datastore(ds["store"]).namespace.get()
            for ns in sorted(namespaces, key=lambda x: x["ns"]):
                group = self._pbs.admin.datastore(ds["store"]).groups.get(ns=ns["ns"])
                gname = self._get_group_name(ds["store"], ns["ns"])
                groups[gname] = group

        print(json.dumps(groups))

    def get_snapshots(self):
        datastore_usage = self._get_datastore_usage()
        snapshots = {}
        for ds in sorted(datastore_usage, key=lambda x: x["store"]):
            if ds["store"] in self._args.exclude:
                continue
            namespaces = self._pbs.admin.datastore(ds["store"]).namespace.get()
            for ns in sorted(namespaces, key=lambda x: x["ns"]):
                snapshot = self._pbs.admin.datastore(ds["store"]).snapshots.get(ns=ns["ns"])
                gname = self._get_group_name(ds["store"], ns["ns"])
                snapshots[gname] = snapshot

        print(json.dumps(snapshots))

    def get_datastores(self):
        datastore_usage = self._get_datastore_usage()
        for d in datastore_usage:
            del d['history']
        print(json.dumps(datastore_usage))

    def run(self):
        if self._args.mode not in self._modes:
            print("Invalid mode", file=sys.stderr)
            self._parser.print_help()
            exit(1)

        self._modes[self._args.mode]()


if __name__ == "__main__":
    QueryPBS().run()
