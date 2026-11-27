# Impacket - Collection of Python classes for working with network protocols.
#
# Copyright Fortra, LLC and its affiliated companies
#
# All rights reserved.
#
# This software is provided under a slightly modified version
# of the Apache Software License. See the accompanying LICENSE file
# for more information.
#
# Description:
#   A similar approach to smbexec but executing commands through WMI.
#   Main advantage here is it runs under the user (has to be Admin)
#   account, not SYSTEM, plus, it doesn't generate noisy messages
#   in the event log that smbexec.py does when creating a service.
#   Drawback is it needs DCOM, hence, I have to be able to access
#   DCOM ports at the target machine.
#
# Author:
#   beto (@agsolino)
#
# Reference for:
#   DCOM
#

from __future__ import division
from __future__ import print_function

import argparse
import codecs
import hmac
import io
import json
import logging
import math
import ntpath
import os
import select
import socket
import sys
import threading
import time
import uuid
from contextlib import contextmanager
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from base64 import b64encode
from impacket import version
from impacket.dcerpc.v5.dcom import wmi
from impacket.dcerpc.v5.dcomrt import (
    DCOMConnection,
    COMVERSION,
    INTERFACE,
    IObjectExporter,
)
from impacket.dcerpc.v5.dtypes import NULL
from impacket.examples import logger
from impacket.examples.utils import parse_target
from impacket.krb5.keytab import Keytab
from impacket.smbconnection import (
    SMBConnection,
    SMB_DIALECT,
    SMB2_DIALECT_002,
    SMB2_DIALECT_21,
)
from six import PY2

import cmd

CODEC = sys.stdout.encoding or "utf-8"
ENV_KEY_SERVER_TOKEN = "WMI_SERVER_TOKEN"
REQUEST_LOCAL = threading.local()

# setting environment variable
ENV_KEY_WMI_PASSWORD = "WMI_PASSWORD"


class WMIEXEC:
    def __init__(
        self,
        command="",
        username="",
        password="",
        domain="",
        hashes=None,
        aesKey=None,
        share=None,
        noOutput=False,
        doKerberos=False,
        kdcHost=None,
        remoteHost="",
        shell_type=None,
        stdout=None,
        log=None,
        codec=None,
        request=None,
    ):
        self.stdout = stdout or sys.stdout
        self.log = log or logging.getLogger()
        self.codec = codec or CODEC
        self.request = request
        self.__command = command
        self.__username = username
        self.__password = password
        self.__domain = domain
        self.__lmhash = ""
        self.__nthash = ""
        self.__aesKey = aesKey
        self.__share = share
        self.__noOutput = noOutput
        self.__doKerberos = doKerberos
        self.__kdcHost = kdcHost
        self.__remoteHost = remoteHost
        self.__shell_type = shell_type
        self.shell = None
        if hashes is not None:
            self.__lmhash, self.__nthash = hashes.split(":")

    def run(self, addr, silentCommand=False):
        smbConnection = None
        dcom = None
        try:
            self._run(addr, silentCommand)
        finally:
            # Cleanup never sends a logoff RPC after cancellation.
            smbConnection = getattr(self, "_smbConnection", None)
            dcom = getattr(self, "_dcom", None)
            if self.request is None:
                if smbConnection is not None:
                    smbConnection.logoff()
                if dcom is not None:
                    dcom.disconnect()
            else:
                self.request.close_sockets()
                if dcom is not None:
                    dcom.disconnect()

    def _run(self, addr, silentCommand):
        if self.__noOutput is False and silentCommand is False:
            smbConnection = SMBConnection(
                addr,
                self.__remoteHost,
                timeout=self.request.remaining() if self.request else 60,
            )
            self._smbConnection = smbConnection
            if self.__doKerberos is False:
                smbConnection.login(
                    self.__username,
                    self.__password,
                    self.__domain,
                    self.__lmhash,
                    self.__nthash,
                )
            else:
                smbConnection.kerberosLogin(
                    self.__username,
                    self.__password,
                    self.__domain,
                    self.__lmhash,
                    self.__nthash,
                    self.__aesKey,
                    kdcHost=self.__kdcHost,
                )

            dialect = smbConnection.getDialect()
        #             if dialect == SMB_DIALECT:
        #                 logging.info("SMBv1 dialect used")
        #             elif dialect == SMB2_DIALECT_002:
        #                 logging.info("SMBv2.0 dialect used")
        #             elif dialect == SMB2_DIALECT_21:
        #                 logging.info("SMBv2.1 dialect used")
        #             else:
        #                 logging.info("SMBv3.0 dialect used")
        else:
            smbConnection = None

        connection_type = RequestDCOMConnection if self.request else DCOMConnection
        dcom = connection_type(
            addr,
            self.__username,
            self.__password,
            self.__domain,
            self.__lmhash,
            self.__nthash,
            self.__aesKey,
            oxidResolver=True,
            doKerberos=self.__doKerberos,
            kdcHost=self.__kdcHost,
            remoteHost=self.__remoteHost,
        )
        self._dcom = dcom
        try:
            iInterface = dcom.CoCreateInstanceEx(
                wmi.CLSID_WbemLevel1Login, wmi.IID_IWbemLevel1Login
            )
            iWbemLevel1Login = wmi.IWbemLevel1Login(iInterface)
            iWbemServices = iWbemLevel1Login.NTLMLogin("//./root/cimv2", NULL, NULL)
            iWbemLevel1Login.RemRelease()

            win32Process, _ = iWbemServices.GetObject("Win32_Process")

            self.shell = RemoteShell(
                self.__share,
                win32Process,
                smbConnection,
                self.__shell_type,
                silentCommand,
                stdout=self.stdout,
                log=self.log,
                codec=self.codec,
                request=self.request,
            )
            if self.__command != " ":
                self.shell.onecmd(self.__command)
            else:
                self.shell.cmdloop()
        except (Exception, KeyboardInterrupt):
            if self.request is not None:
                raise
            if self.log.isEnabledFor(logging.DEBUG):
                import traceback

                traceback.print_exc()
            raise


class RemoteShell(cmd.Cmd):
    def __init__(
        self,
        share,
        win32Process,
        smbConnection,
        shell_type,
        silentCommand=False,
        stdout=None,
        log=None,
        codec=None,
        request=None,
    ):
        cmd.Cmd.__init__(self, stdout=stdout)
        self.log = log or logging.getLogger()
        self.codec = codec or CODEC
        self.request = request
        self.__localDirectory = os.getcwd()
        self.__share = share
        self.__output = ""
        self.__outputBuffer = str("")
        self.__shell = "cmd.exe /Q /c "
        self.__shell_type = shell_type
        self.__pwsh = "powershell.exe -NoP -NoL -sta -NonI -W Hidden -Exec Bypass -Enc "
        self.__win32Process = win32Process
        self.__transferClient = smbConnection
        self.__silentCommand = silentCommand
        self.__pwd = str("C:\\")
        self.__noOutput = False
        self.intro = "[!] Launching semi-interactive shell - Careful what you execute\n[!] Press help for extra shell commands"

        # We don't wanna deal with timeouts from now on.
        if self.__transferClient is not None:
            self.__transferClient.setTimeout(
                self.request.remaining() if self.request else 100000
            )
            self.do_cd("\\")
        else:
            self.__noOutput = True

        # If the user wants to just execute a command without cmd.exe, set raw command and set no output
        if self.__silentCommand is True:
            self.__shell = ""

    def do_shell(self, s):
        if self.request:
            raise ValueError("local shell commands are not supported in server mode")
        os.system(s)

    def do_help(self, line):
        print(
            """
 lcd {path}                 - changes the current local directory to {path}
 exit                       - terminates the server process (and this session)
 lput {src_file, dst_path}   - uploads a local file to the dst_path (dst_path = default current directory)
 lget {file}                 - downloads pathname to the current local dir
 ! {cmd}                    - executes a local shell cmd
""",
            file=self.stdout,
        )

    def do_lcd(self, s):
        if s == "":
            print(self.__localDirectory, file=self.stdout)
        else:
            try:
                if self.request:
                    path = os.path.abspath(os.path.join(self.__localDirectory, s))
                    if not os.path.isdir(path):
                        raise ValueError("local directory does not exist")
                    self.__localDirectory = path
                else:
                    os.chdir(s)
                    self.__localDirectory = os.getcwd()
            except Exception as e:
                self.log.error(str(e))

    def do_lget(self, src_path):
        filename = None
        try:
            import ntpath

            newPath = ntpath.normpath(ntpath.join(self.__pwd, src_path))
            drive, tail = ntpath.splitdrive(newPath)
            filename = os.path.join(self.__localDirectory, ntpath.basename(tail))
            self.log.info("Downloading %s\\%s" % (drive, tail))
            with open(filename, "wb") as fh:
                self.__transferClient.getFile(drive[:-1] + "$", tail, fh.write)

        except Exception as e:
            if filename is not None and os.path.exists(filename):
                os.remove(filename)
            if self.request:
                raise
            self.log.error(str(e))

    def do_lput(self, s):
        try:
            params = s.split(" ")
            if len(params) > 1:
                src_path = params[0]
                dst_path = params[1]
            elif len(params) == 1:
                src_path = params[0]
                dst_path = ""

            src_file = os.path.basename(src_path)
            src_path = os.path.join(self.__localDirectory, src_path)
            dst_path = dst_path.replace("/", "\\")
            import ntpath

            pathname = ntpath.join(ntpath.join(self.__pwd, dst_path), src_file)
            drive, tail = ntpath.splitdrive(pathname)
            self.log.info("Uploading %s to %s" % (src_file, pathname))
            with open(src_path, "rb") as fh:
                self.__transferClient.putFile(drive[:-1] + "$", tail, fh.read)
        except Exception as e:
            if self.request:
                raise
            self.log.critical(str(e))

    def do_exit(self, s):
        return True

    def do_EOF(self, s):
        print(file=self.stdout)
        return self.do_exit(s)

    def emptyline(self):
        return False

    def do_cd(self, s):
        self.execute_remote("cd " + s)
        if len(self.__outputBuffer.strip("\r\n")) > 0:
            print(self.__outputBuffer, file=self.stdout)
            self.__outputBuffer = ""
        else:
            if PY2:
                self.__pwd = ntpath.normpath(
                    ntpath.join(self.__pwd, s.decode(sys.stdin.encoding))
                )
            else:
                self.__pwd = ntpath.normpath(ntpath.join(self.__pwd, s))
            self.execute_remote("cd ")
            self.__pwd = self.__outputBuffer.strip("\r\n")
            self.prompt = self.__pwd + ">"
            if self.__shell_type == "powershell":
                self.prompt = "PS " + self.prompt + " "
            self.__outputBuffer = ""

    def default(self, line):
        # Let's try to guess if the user is trying to change drive
        if len(line) == 2 and line[1] == ":":
            # Execute the command and see if the drive is valid
            self.execute_remote(line)
            if len(self.__outputBuffer.strip("\r\n")) > 0:
                # Something went wrong
                print(self.__outputBuffer, file=self.stdout)
                self.__outputBuffer = ""
            else:
                # Drive valid, now we should get the current path
                self.__pwd = line
                self.execute_remote("cd ")
                self.__pwd = self.__outputBuffer.strip("\r\n")
                self.prompt = self.__pwd + ">"
                self.__outputBuffer = ""
        else:
            if line != "":
                self.send_data(line)

    def get_output(self):
        def output_callback(data):
            try:
                self.__outputBuffer += data.decode(self.codec)
            except UnicodeDecodeError:
                self.log.error(
                    "Decoding error detected, consider running chcp.com at the target,\nmap the result with "
                    "https://docs.python.org/3/library/codecs.html#standard-encodings\nand then execute wmiexec.py "
                    "again with -codec and the corresponding codec"
                )
                self.__outputBuffer += data.decode(self.codec, errors="replace")

        if self.__noOutput is True:
            self.__outputBuffer = ""
            return

        while True:
            if self.request:
                self.request.check()
            try:
                self.__transferClient.getFile(
                    self.__share, self.__output, output_callback
                )
                break
            except Exception as e:
                if any(
                    status in str(e)
                    for status in (
                        "STATUS_SHARING_VIOLATION",
                        "STATUS_OBJECT_NAME_NOT_FOUND",
                        "STATUS_NO_SUCH_FILE",
                    )
                ):
                    # Output may not exist yet, or the remote process still owns it.
                    if self.request:
                        self.request.cancelled.wait(1)
                    else:
                        time.sleep(1)
                elif str(e).find("Broken") >= 0 and self.request is None:
                    # Only the legacy CLI reconnects; server requests never replay.
                    self.log.debug("Connection broken, trying to recreate it")
                    self.__transferClient.reconnect()
                    return self.get_output()
                else:
                    raise
        self.__transferClient.deleteFile(self.__share, self.__output)

    def execute_remote(self, data, shell_type="cmd"):
        if self.request:
            self.request.check()
        self.__output = "\\__" + uuid.uuid4().hex
        if shell_type == "powershell":
            data = '$ProgressPreference="SilentlyContinue";' + data
            data = self.__pwsh + b64encode(data.encode("utf-16le")).decode()

        command = self.__shell + data

        if self.__noOutput is False:
            command += (
                " 1> " + "\\\\127.0.0.1\\%s" % self.__share + self.__output + " 2>&1"
            )
        if PY2:
            self.__win32Process.Create(
                command.decode(sys.stdin.encoding), self.__pwd, None
            )
        else:
            self.__win32Process.Create(command, self.__pwd, None)
        self.get_output()

    def send_data(self, data):
        self.execute_remote(data, self.__shell_type)
        print(self.__outputBuffer, file=self.stdout)
        self.__outputBuffer = ""


class AuthFileSyntaxError(Exception):
    """raised by load_smbclient_auth_file if it encounters a syntax error
    while loading the smbclient-style authentication file."""

    def __init__(self, path, lineno, reason):
        self.path = path
        self.lineno = lineno
        self.reason = reason

    def __str__(self):
        return "Syntax error in auth file %s line %d: %s" % (
            self.path,
            self.lineno,
            self.reason,
        )


def load_smbclient_auth_file(path):
    """Load credentials from an smbclient-style authentication file (used by
    smbclient, mount.cifs and others).  returns (domain, username, password)
    or raises AuthFileSyntaxError or any I/O exceptions."""

    lineno = 0
    domain = None
    username = None
    password = None
    for line in open(path):
        lineno += 1

        line = line.strip()

        if line.startswith("#") or line == "":
            continue

        parts = line.split("=", 1)
        if len(parts) != 2:
            raise AuthFileSyntaxError(path, lineno, 'No "=" present in line')

        (k, v) = (parts[0].strip(), parts[1].strip())

        if k == "username":
            username = v
        elif k == "password":
            password = v
        elif k == "domain":
            domain = v
        else:
            raise AuthFileSyntaxError(path, lineno, "Unknown option %s" % repr(k))

    return (domain, username, password)


class RequestContext:
    def __init__(self, connection, timeout):
        self.connection = connection
        self.deadline = time.monotonic() + timeout if timeout else None
        self.cancelled = threading.Event()
        self.finished = threading.Event()
        self.sockets = set()
        self.closing = False
        self.lock = threading.Lock()
        self.error = "request cancelled"

    def check(self):
        if self.deadline is not None and time.monotonic() >= self.deadline:
            self.cancel("request timeout")
        if self.cancelled.is_set():
            raise TimeoutError(self.error)

    def remaining(self):
        self.check()
        return max(0.001, self.deadline - time.monotonic()) if self.deadline else 60

    def add_socket(self, sock):
        with self.lock:
            if self.closing:
                sock.close()
                raise TimeoutError("request connection is closing")
            self.sockets.add(sock)
        self.check()

    def close_sockets(self):
        with self.lock:
            self.closing = True
            sockets = list(self.sockets)
        for sock in sockets:
            try:
                sock.shutdown(socket.SHUT_RDWR)
            except OSError:
                pass
            sock.close()

    def cancel(self, error):
        self.error = error
        self.cancelled.set()
        self.close_sockets()

    def watch(self):
        while not self.finished.wait(0.05):
            if self.deadline is not None and time.monotonic() >= self.deadline:
                self.cancel("request timeout")
            try:
                readable, _, _ = select.select([self.connection], [], [], 0)
                if readable and not self.connection.recv(1, socket.MSG_PEEK):
                    self.cancel("request disconnected")
            except OSError:
                self.cancel("request disconnected")
            if self.cancelled.is_set():
                # Catch sockets created by a connect already in progress as well.
                self.close_sockets()


class RequestLogHandler(logging.Handler):
    def emit(self, record):
        log = getattr(REQUEST_LOCAL, "log", None)
        if log is not None and record.levelno >= log.level:
            log.handle(record)


def install_request_transport_hooks():
    # Installed once at server startup, never changed by individual requests.
    # Impacket creates transports deep inside DCOM interface binding; intercept
    # their socket boundary to stop blocked I/O and prevent sends after cancel.
    from impacket import nmb
    from impacket.dcerpc.v5 import transport

    tcp_connect = transport.TCPTransport.connect
    tcp_send = transport.TCPTransport.send
    smb_connect = nmb.NetBIOSTCPSession._setup_connection
    smb_send = nmb.NetBIOSTCPSession.send_packet

    def connect_tcp(self):
        request = getattr(REQUEST_LOCAL, "request", None)
        if request is None:
            return tcp_connect(self)
        request.check()
        af, kind, proto, _, address = socket.getaddrinfo(
            self.getRemoteHost(), self.get_dport(), 0, socket.SOCK_STREAM
        )[0]
        request.check()
        sock = socket.socket(af, kind, proto)
        self._TCPTransport__socket = sock
        request.add_socket(sock)
        sock.settimeout(request.remaining())
        sock.connect(address)
        request.check()
        return 1

    def connect_smb(self, peer, timeout=None):
        request = getattr(REQUEST_LOCAL, "request", None)
        if request is None:
            return smb_connect(self, peer, timeout)
        request.check()
        af, kind, proto, _, address = socket.getaddrinfo(
            peer[0], peer[1], 0, socket.SOCK_STREAM
        )[0]
        request.check()
        sock = socket.socket(af, kind, proto)
        request.add_socket(sock)
        sock.settimeout(request.remaining())
        sock.connect(address)
        request.check()
        return sock

    def send_tcp(self, *args, **kwargs):
        request = getattr(REQUEST_LOCAL, "request", None)
        if request:
            request.check()
        return tcp_send(self, *args, **kwargs)

    def send_smb(self, *args, **kwargs):
        request = getattr(REQUEST_LOCAL, "request", None)
        if request:
            request.check()
        return smb_send(self, *args, **kwargs)

    transport.TCPTransport.connect = connect_tcp
    transport.TCPTransport.send = send_tcp
    nmb.NetBIOSTCPSession._setup_connection = connect_smb
    nmb.NetBIOSTCPSession.send_packet = send_smb
    dependency_log = logging.getLogger("impacket")
    dependency_log.setLevel(logging.DEBUG)
    dependency_log.propagate = False
    dependency_log.addHandler(RequestLogHandler())
    logging.getLogger().setLevel(logging.DEBUG)
    logging.getLogger().addHandler(RequestLogHandler())


class RequestDCOMConnection(DCOMConnection):
    def __init__(self, target, *args, **kwargs):
        self.target = target
        self.request = REQUEST_LOCAL.request
        self.log = REQUEST_LOCAL.log
        self.stopped = threading.Event()
        self.pinger = None
        super().__init__(target, *args, **kwargs)

    def initTimer(self):
        # Impacket's global timer iterates mutable maps of every target without
        # locks. Each serialized target instead owns its keepalive and cleanup.
        if self.pinger is None:
            self.pinger = threading.Thread(target=self.ping, daemon=True)
            self.pinger.start()

    def ping(self):
        REQUEST_LOCAL.request = self.request
        REQUEST_LOCAL.log = self.log
        try:
            while not self.stopped.wait(120):
                try:
                    self.request.check()
                    state = DCOMConnection.OID_SET.get(self.target)
                    if state is None:
                        continue
                    added = DCOMConnection.OID_ADD.pop(self.target, set())
                    deleted = DCOMConnection.OID_DEL.pop(self.target, set())
                    exporter = IObjectExporter(DCOMConnection.PORTMAPS[self.target])
                    if added or deleted:
                        response = exporter.ComplexPing(
                            state["setid"], 0, added, deleted
                        )
                        state["oids"] -= deleted
                        state["oids"] |= added
                        state["setid"] = response["pSetId"]
                    else:
                        exporter.SimplePing(state["setid"])
                except Exception as e:
                    if self.stopped.is_set() or self.request.cancelled.is_set():
                        break
                    self.log.error("DCOM keepalive failed: %s", e)
        finally:
            del REQUEST_LOCAL.request
            del REQUEST_LOCAL.log

    def disconnect(self):
        self.stopped.set()
        if self.pinger is not None:
            self.pinger.join()
        clear_target_connections(self.target)


def clear_target_connections(target):
    # Called while holding the target lock, after all registered sockets close.
    for cache in (
        DCOMConnection.PORTMAPS,
        DCOMConnection.OID_SET,
        DCOMConnection.OID_ADD,
        DCOMConnection.OID_DEL,
        INTERFACE.CONNECTIONS,
    ):
        cache.pop(target, None)


class RequestArgumentParser(argparse.ArgumentParser):
    def error(self, message):
        raise ValueError(message)

    def exit(self, status=0, message=None):
        raise ValueError(message or "help is not supported in server mode")

    def _print_message(self, message, file=None):
        # argparse help must not leak into the server's port announcement pipe.
        pass


def make_parser(server=False):
    parser_type = RequestArgumentParser if server else argparse.ArgumentParser
    parser = parser_type(
        add_help=True,
        description="Executes a semi-interactive shell using Windows "
        "Management Instrumentation.",
    )
    parser.add_argument(
        "target",
        action="store",
        help="[[domain/]username[:password]@]<targetName or address>",
    )
    parser.add_argument(
        "-share",
        action="store",
        default="ADMIN$",
        help="share where the output will be grabbed from (default ADMIN$)",
    )
    parser.add_argument(
        "-nooutput",
        action="store_true",
        default=False,
        help="whether or not to print the output (no SMB connection created)",
    )
    parser.add_argument(
        "-ts", action="store_true", help="Adds timestamp to every logging output"
    )
    parser.add_argument(
        "-silentcommand",
        action="store_true",
        default=False,
        help="does not execute cmd.exe to run given command (no output)",
    )
    parser.add_argument("-debug", action="store_true", help="Turn DEBUG output ON")
    parser.add_argument(
        "-codec",
        action="store",
        help="Sets encoding used (codec) from the target's output (default "
        '"%s"). If errors are detected, run chcp.com at the target, '
        "map the result with "
        "https://docs.python.org/3/library/codecs.html#standard-encodings and then execute wmiexec.py "
        "again with -codec and the corresponding codec " % CODEC,
    )
    parser.add_argument(
        "-shell-type",
        action="store",
        default="cmd",
        choices=["cmd", "powershell"],
        help="choose a command processor for the semi-interactive shell",
    )
    parser.add_argument(
        "-com-version",
        action="store",
        metavar="MAJOR_VERSION:MINOR_VERSION",
        help="DCOM version, format is MAJOR_VERSION:MINOR_VERSION e.g. 5.7",
    )
    parser.add_argument(
        "command",
        nargs="*",
        default=" ",
        help="command to execute at the target. If empty it will "
        "launch a semi-interactive shell",
    )

    group = parser.add_argument_group("authentication")

    group.add_argument(
        "-hashes",
        action="store",
        metavar="LMHASH:NTHASH",
        help="NTLM hashes, format is LMHASH:NTHASH",
    )
    group.add_argument(
        "-no-pass", action="store_true", help="don't ask for password (useful for -k)"
    )
    group.add_argument(
        "-k",
        action="store_true",
        help="Use Kerberos authentication. Grabs credentials from ccache file "
        "(KRB5CCNAME) based on target parameters. If valid credentials cannot be found, it will use the "
        "ones specified in the command line",
    )
    group.add_argument(
        "-aesKey",
        action="store",
        metavar="hex key",
        help="AES key to use for Kerberos Authentication (128 or 256 bits)",
    )
    group.add_argument(
        "-dc-ip",
        action="store",
        metavar="ip address",
        help="IP Address of the domain controller. If "
        "ommited it use the domain part (FQDN) specified in the target parameter",
    )
    group.add_argument(
        "-target-ip",
        action="store",
        metavar="ip address",
        help="IP Address of the target machine. If omitted it will use whatever was specified as target. "
        "This is useful when target is the NetBIOS name and you cannot resolve it",
    )
    group.add_argument(
        "-A",
        action="store",
        metavar="authfile",
        help="smbclient/mount.cifs-style authentication file. "
        "See smbclient man page's -A option.",
    )
    group.add_argument(
        "-keytab", action="store", help="Read keys for SPN from keytab file"
    )

    return parser


def execute_request(payload, context, server):
    output = io.StringIO()
    errors = io.StringIO()
    log = logging.Logger("wmiexec-request")
    handler = logging.StreamHandler(errors)
    log.addHandler(handler)
    error = ""
    REQUEST_LOCAL.request = context
    REQUEST_LOCAL.log = log
    try:
        args = payload["args"]
        password = payload["password"]
        if not isinstance(args, list) or not all(isinstance(arg, str) for arg in args):
            raise ValueError("args must be an array of strings")
        if not isinstance(password, str):
            raise ValueError("password must be a string")
        options = make_parser(server=True).parse_args(args)
        log.setLevel(logging.DEBUG if options.debug else logging.INFO)
        if options.ts:
            handler.setFormatter(logging.Formatter("%(asctime)s %(message)s"))
        codec = options.codec or CODEC
        codecs.lookup(codec)
        command = " ".join(options.command)
        if not command.strip():
            raise ValueError("interactive shell is not supported in server mode")
        if options.com_version and options.com_version != "5.7":
            raise ValueError("server mode only supports the default DCOM version 5.7")
        domain, username, target_password, address = parse_target(options.target)
        if options.A:
            domain, username, target_password = load_smbclient_auth_file(options.A)
        if options.keytab:
            Keytab.loadKeysFromKeytab(options.keytab, username, domain, options)
            options.k = True
        if options.aesKey:
            options.k = True
        target_ip = options.target_ip or address
        # PORTMAPS and object caches in Impacket are keyed by the target name,
        # not the thread or credentials. Only requests for that key serialize.
        with server.target_lock(address.casefold(), context):
            context.check()
            try:
                executer = WMIEXEC(
                    command,
                    username,
                    target_password or password,
                    domain or "",
                    options.hashes,
                    options.aesKey,
                    options.share,
                    options.nooutput,
                    options.k,
                    options.dc_ip,
                    target_ip,
                    options.shell_type,
                    stdout=output,
                    log=log,
                    codec=codec,
                    request=context,
                )
                executer.run(address, options.silentcommand)
            finally:
                context.close_sockets()
                clear_target_connections(address)
    except Exception as e:
        error = context.error if context.cancelled.is_set() else str(e)
    finally:
        del REQUEST_LOCAL.request
        del REQUEST_LOCAL.log
    return {"stdout": output.getvalue(), "stderr": errors.getvalue(), "error": error}


class ExecuteServer(ThreadingHTTPServer):
    request_queue_size = 300

    def __init__(self, address, handler):
        super().__init__(address, handler)
        self.targets = {}
        self.targets_lock = threading.Lock()

    @contextmanager
    def target_lock(self, target, context):
        with self.targets_lock:
            lock, users = self.targets.get(target, (threading.Lock(), 0))
            self.targets[target] = (lock, users + 1)
        acquired = False
        try:
            while not acquired:
                context.check()
                acquired = lock.acquire(timeout=0.05)
            yield
        finally:
            if acquired:
                lock.release()
            with self.targets_lock:
                _, users = self.targets[target]
                if users == 1:
                    del self.targets[target]
                else:
                    self.targets[target] = (lock, users - 1)


class ExecuteHandler(BaseHTTPRequestHandler):
    protocol_version = "HTTP/1.1"

    def log_message(self, *args):
        pass

    def setup(self):
        self.request.setsockopt(socket.IPPROTO_TCP, socket.TCP_NODELAY, 1)
        super().setup()
        self.connection.settimeout(60)

    def authorized(self):
        return hmac.compare_digest(
            self.headers.get("Authorization", "").encode("utf-8"), self.server.token
        )

    def respond(self, status, data):
        body = json.dumps(data).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def do_GET(self):
        if not self.authorized():
            self.send_error(403)
            return
        if self.path != "/healthz":
            self.send_error(404)
            return
        self.respond(200, {})

    def do_POST(self):
        if not self.authorized():
            self.send_error(403)
            return
        if self.path != "/execute":
            self.send_error(404)
            return
        try:
            length = int(self.headers["Content-Length"])
            if length < 0:
                raise ValueError("invalid content length")
            payload = json.loads(self.rfile.read(length))
            timeout = payload["timeout"]
            if isinstance(timeout, bool) or not isinstance(timeout, (int, float)):
                raise ValueError("timeout must be a number")
            if timeout < 0 or not math.isfinite(timeout):
                raise ValueError("invalid timeout")
            if "args" not in payload or "password" not in payload:
                raise ValueError("missing arguments")
        except (ValueError, TypeError, KeyError):
            self.send_error(400)
            return
        context = RequestContext(self.connection, timeout)
        watcher = threading.Thread(target=context.watch, daemon=True)
        watcher.start()
        try:
            result = execute_request(payload, context, self.server)
        finally:
            context.finished.set()
            watcher.join()
        try:
            self.respond(200, result)
        except (BrokenPipeError, ConnectionResetError):
            self.close_connection = True


def serve(start_port, end_port):
    token = os.environ.pop(ENV_KEY_SERVER_TOKEN).encode("utf-8")
    if not token:
        raise ValueError("missing server token")
    install_request_transport_hooks()
    for port in range(start_port, end_port + 1):
        try:
            server = ExecuteServer(("127.0.0.1", port), ExecuteHandler)
        except OSError:
            continue
        with server:
            server.token = token

            def wait_for_parent():
                sys.stdin.buffer.read()
                server.shutdown()

            threading.Thread(target=wait_for_parent, daemon=True).start()
            print(port, flush=True)
            server.serve_forever()
        return
    raise RuntimeError("no available port in configured range")


def main():
    global CODEC
    if len(sys.argv) > 1 and sys.argv[1] == "--serve":
        parser = argparse.ArgumentParser()
        parser.add_argument("--serve", nargs=2, type=int)
        serve(*parser.parse_args().serve)
        return

    print(version.BANNER)
    parser = make_parser()
    if len(sys.argv) == 1:
        parser.print_help()
        sys.exit(1)
    options = parser.parse_args()

    # Init the example's logger theme
    logger.init(options.ts)

    if options.codec is not None:
        CODEC = options.codec
    else:
        if CODEC is None:
            CODEC = "utf-8"

    if " ".join(options.command) == " " and options.nooutput is True:
        logging.error("-nooutput switch and interactive shell not supported")
        sys.exit(1)
    if options.silentcommand and options.command == " ":
        logging.error("-silentcommand switch and interactive shell not supported")
        sys.exit(1)

    if options.debug is True:
        logging.getLogger().setLevel(logging.DEBUG)
        # Print the Library's installation path
        logging.debug(version.getInstallationPath())
    else:
        logging.getLogger().setLevel(logging.INFO)

    if options.com_version is not None:
        try:
            major_version, minor_version = options.com_version.split(".")
            COMVERSION.set_default_version(int(major_version), int(minor_version))
        except Exception:
            logging.error(
                'Wrong COMVERSION format, use dot separated integers e.g. "5.7"'
            )
            sys.exit(1)

    domain, username, password, address = parse_target(options.target)

    try:
        if options.A is not None:
            (domain, username, password) = load_smbclient_auth_file(options.A)
            logging.debug(
                "loaded smbclient auth file: domain=%s, username=%s, password=%s"
                % (repr(domain), repr(username), repr(password))
            )

        if options.target_ip is None:
            options.target_ip = address

        if domain is None:
            domain = ""

        if options.keytab is not None:
            Keytab.loadKeysFromKeytab(options.keytab, username, domain, options)
            options.k = True

        password = password if password else os.environ.get(ENV_KEY_WMI_PASSWORD)

        if (
            password == ""
            and username != ""
            and options.hashes is None
            and options.no_pass is False
            and options.aesKey is None
        ):
            from getpass import getpass

            password = getpass("Password:")

        if options.aesKey is not None:
            options.k = True

        executer = WMIEXEC(
            " ".join(options.command),
            username,
            password,
            domain,
            options.hashes,
            options.aesKey,
            options.share,
            options.nooutput,
            options.k,
            options.dc_ip,
            options.target_ip,
            options.shell_type,
        )
        executer.run(address, options.silentcommand)
    except KeyboardInterrupt as e:
        logging.error(str(e))
    except Exception as e:
        if logging.getLogger().level == logging.DEBUG:
            import traceback

            traceback.print_exc()
        logging.error(str(e))
        sys.exit(1)

    sys.exit(0)


if __name__ == "__main__":
    main()
