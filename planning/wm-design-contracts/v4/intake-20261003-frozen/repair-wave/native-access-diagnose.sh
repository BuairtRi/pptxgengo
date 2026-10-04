#!/bin/bash
# Read-only comparison: run the same file in Terminal and the agent tool session.
# Prints JSON; does not open/close applications or change macOS permissions.
set -u
python3 - <<'PY'
import ctypes, datetime, json, os, subprocess

def command(args, timeout=10):
    try:
        p = subprocess.run(args, capture_output=True, text=True, timeout=timeout)
        return {'argv': args, 'exit_code': p.returncode, 'stdout': p.stdout.strip()[:1200], 'stderr': p.stderr.strip()[:1200]}
    except Exception as e:
        return {'argv': args, 'error': str(e)}

result = {'schema': 'pptxgengo.native-access-diagnostic.v1',
          'observed_at': datetime.datetime.now(datetime.timezone.utc).isoformat(),
          'uid': os.getuid(), 'pid': os.getpid(), 'ppid': os.getppid(),
          'environment': {k: os.environ.get(k, '') for k in ('HOME', 'TMPDIR', 'XPC_SERVICE_NAME', 'XPC_FLAGS', 'SECURITYSESSIONID')},
          'launchd': [command(['/bin/launchctl', 'manageruid']), command(['/bin/launchctl', 'managername'])]}
ancestors = []
pid = os.getpid()
for _ in range(5):
    p = command(['/bin/ps', '-p', str(pid), '-o', 'pid=,ppid=,uid=,comm='])
    ancestors.append(p)
    fields = p.get('stdout', '').split(None, 3)
    if len(fields) < 2 or fields[1] == '0': break
    pid = int(fields[1])
result['ancestors'] = ancestors
result['applescript_arithmetic'] = command(['/usr/bin/osascript', '-e', '1 + 1'])
# Query only when the executable is already running, to avoid starting the app.
processes = subprocess.run(['/bin/ps', '-axo', 'pid=,uid=,comm='], capture_output=True, text=True, timeout=10)
pp_rows = []
for line in processes.stdout.splitlines():
    fields = line.split(None, 2)
    if len(fields) == 3 and fields[2] == '/Applications/Microsoft PowerPoint.app/Contents/MacOS/Microsoft PowerPoint':
        pp_rows.append({'pid': int(fields[0]), 'uid': int(fields[1]), 'executable': fields[2]})
result['powerpoint_running'] = pp_rows
if any(row['uid'] == os.getuid() for row in pp_rows):
    result['powerpoint_count'] = command(['/usr/bin/osascript', '-e', 'tell application "/Applications/Microsoft PowerPoint.app" to count presentations'])
else:
    result['powerpoint_count'] = {'skipped': 'PowerPoint is not running for this UID; this diagnostic never starts it.'}
try:
    lib = ctypes.CDLL('/usr/lib/libSystem.B.dylib')
    task = ctypes.c_uint32.in_dll(lib, 'mach_task_self_').value
    bootstrap = ctypes.c_uint32.in_dll(lib, 'bootstrap_port').value
    lib.task_get_special_port.argtypes = [ctypes.c_uint32, ctypes.c_int, ctypes.POINTER(ctypes.c_uint32)]
    lib.task_get_special_port.restype = ctypes.c_int
    lib.mach_error_string.argtypes = [ctypes.c_int]
    lib.mach_error_string.restype = ctypes.c_char_p
    lib.bootstrap_look_up.argtypes = [ctypes.c_uint32, ctypes.c_char_p, ctypes.POINTER(ctypes.c_uint32)]
    lib.bootstrap_look_up.restype = ctypes.c_int
    port = ctypes.c_uint32()
    status = lib.task_get_special_port(task, 4, ctypes.byref(port))
    probe = {'task_self': task, 'global_bootstrap_port': bootstrap,
             'task_get_special_port_status': status, 'task_bootstrap_port': port.value,
             'task_bootstrap_dead': port.value == 0xffffffff, 'task_bootstrap_null': port.value == 0}
    if status == 0:
        destination = ctypes.c_uint32()
        status = lib.bootstrap_look_up(port.value, b'com.apple.coreservices.launchservicesd', ctypes.byref(destination))
        probe['launchservices_lookup'] = {'status': status, 'description': lib.mach_error_string(status).decode(), 'port': destination.value}
        if destination.value not in (0, 0xffffffff): lib.mach_port_deallocate(task, destination.value)
        if port.value not in (0, 0xffffffff): lib.mach_port_deallocate(task, port.value)
    result['mach_bootstrap'] = probe
except Exception as e:
    result['mach_bootstrap'] = {'error': str(e)}
print(json.dumps(result, indent=2))
PY
