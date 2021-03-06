#!/usr/bin/env python3
# -*- coding: utf-8 -*-
import argparse, sys, subprocess, textwrap, threading, signal

def shell_escape(s):
    return "'" + s.replace("'", "'\"'\"'") + "'"

def run(cmd):
    try:
        subprocess.check_call(cmd, shell=True)
    except subprocess.CalledProcessError as e:
        print(e, file=sys.stderr)

def sig_handler(signum, frame):
    sys.exit(0)

def start(args):
    run_commands = args.run
    pre_commands = args.pre
    post_commands = args.post

    # handing signal to execute finally code.
    signal.signal(signal.SIGTERM, sig_handler)
    signal.signal(signal.SIGINT, sig_handler)

    try:
        # run pre command
        for cmd in pre_commands:
            run(cmd)

        # start run commands
        threads = []
        for cmd in run_commands:
            t = threading.Thread(target=run, args=(cmd,))
            threads.append(t)
            t.start()

        # wait for all run command threads finish
        for t in threads:
            t.join()
    finally:
        # run post command
        for cmd in post_commands:
            run(cmd)

def main():
    parser = argparse.ArgumentParser(
        description="process-starter.py is a utility to start multiple processes",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=textwrap.dedent('''\
            description:
              A utility to start multiple processes

            example:
              process-starter.py --run "your-file-watcher-command" "your-dev-server-start-command"
              process-starter.py --pre "your-build-command" --run "your-dev-server-start-command"

            Copyright (c) Kohki Makimoto <kohki.makimoto@gmail.com>
            The MIT License (MIT)
        '''))

    parser.add_argument("--pre", dest="pre", metavar="COMMAND", nargs='*', help="Set commands that are executed before run commands", default=[])
    parser.add_argument("--post", dest="post", metavar="COMMAND", nargs='*',help="Set commands that are executed after run commands", default=[])
    parser.add_argument("--run", "-r", dest="run", metavar="COMMAND", nargs='*', help="Set commands to run concurrently", default=[])

    if len(sys.argv) == 1:
        parser.print_help()
        sys.exit(1)

    args = parser.parse_args()
    start(args)

if __name__ == '__main__': main()
