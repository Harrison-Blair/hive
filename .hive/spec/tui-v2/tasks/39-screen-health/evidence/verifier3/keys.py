exec(open('.hive/tmp/verify39v3/live.py').read().split("assert call('ping'")[0])
d=root/'.hive/tmp/verify39v3'; evidence=root/'.hive/spec/tui-v2/tasks/39-screen-health/evidence/verifier3'
def run(board,size):
 subprocess.run(['herdr','pane','run',pane,f"HIVE_HOME='{d}/home' COLORTERM=truecolor '{d}/hive' --demo '{root}/.hive/spec/tui-v2/design/{board}.fixture.json' --size {size}"],check=True,stdout=subprocess.DEVNULL);time.sleep(1)
def keys(*k):
 subprocess.run(['herdr','pane','send-keys',pane,*k],check=True,stdout=subprocess.DEVNULL);time.sleep(.5)
def save(name,*wants):
 r=call('pane.read',dict(pane_id=pane,source='visible',format='text'));text=r['result']['read']['text'];(evidence/(name+'.txt')).write_text(text)
 for want in wants:assert want in text,(name,want,text)
run('Health','120x40');keys('down');save('home-down','git');keys('G');save('home-end','gh');keys('enter');save('home-details','[details]');keys('ctrl+c','ctrl+c')
run('HealthAgents','120x40');keys('esc');save('agents-cancel','restart cancelled');keys('j','h');save('agents-hide','h shows');keys('r');save('agents-demo','demo mode: no action');keys('ctrl+c','ctrl+c')
run('CompactHealth','80x24');keys('enter');save('compact-details','esc back','[details]');keys('esc');save('compact-back','[list]');keys('ctrl+c','ctrl+c')
# Records refer only to the owned pane or a nonexistent workspace. No live agent process is used.
state=d/'home/.hive/state/agents';state.mkdir(parents=True,exist_ok=True)
records=[dict(key='coordinator-g1',name='coordinator',role='coordinator',status='live',pane_id=pane,workspace_id='verify39v3-nonexistent',generation=1),dict(key='manager-g1',name='manager',role='manager',status='live'),dict(key='supervisor-g1',name='supervisor',role='supervisor',status='stopped',workspace_id='verify39v3-nonexistent'),dict(key='worker-g1',name='worker',role='worker',status='live')]
for record in records:(state/(record['key']+'.json')).write_text(json.dumps(record))
subprocess.run(['herdr','pane','run',pane,f"HIVE_HOME='{d}/home' COLORTERM=truecolor '{d}/hive' --no-startup"],check=True,stdout=subprocess.DEVNULL);time.sleep(1);keys('5',']','R');save('normal-loaded','coordinator','supervisor','manager','worker');keys('g','r');save('normal-prompt','[restart]');keys('n');save('normal-cancel','restart cancelled');keys('j','j','r');save('normal-manager','restart: coordinator and supervisor only');keys('G','r');save('normal-worker','restart: coordinator and supervisor only');keys('g','r');keys('y');save('normal-confirm-y','workspace_not_found');keys('ctrl+c','ctrl+c')
subprocess.run(['herdr','pane','run',pane,f"HIVE_HOME='{d}/home' COLORTERM=truecolor '{d}/hive' --no-startup"],check=True,stdout=subprocess.DEVNULL);time.sleep(1);keys('5',']','R');keys('g','j','r');save('normal-supervisor','workspace_not_found');keys('ctrl+c','ctrl+c')
records[0]['pane_id']='verify39v3-gone';(state/'coordinator-g1.json').write_text(json.dumps(records[0]))
subprocess.run(['herdr','pane','run',pane,f"HIVE_HOME='{d}/home' COLORTERM=truecolor '{d}/hive' --no-startup"],check=True,stdout=subprocess.DEVNULL);time.sleep(1);keys('5',']','R');keys('g','r');save('normal-stale','workspace_not_found');keys('ctrl+c','ctrl+c')
