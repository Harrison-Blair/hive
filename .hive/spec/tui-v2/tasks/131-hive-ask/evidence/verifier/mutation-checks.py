import pathlib,json,subprocess,re
root=pathlib.Path.cwd();tmp=root/'.hive/tmp/verifier131';ev=root/'.hive/spec/tui-v2/tasks/131-hive-ask/evidence/verifier'
def run(label,changes,packages,pattern):
 mapping={}
 for i,(file,edits) in enumerate(changes.items()):
  data=(root/file).read_text()
  for old,new in edits:
   assert old in data,(label,file,old);data=data.replace(old,new)
  dest=tmp/(label+'-'+str(i)+'.go');dest.write_text(data);mapping[str(root/file)]=str(dest)
 overlay=tmp/(label+'.json');overlay.write_text(json.dumps({'Replace':mapping}))
 args=['go','test','-count=1','-json','-overlay',str(overlay),'-run',pattern]+packages
 with (ev/(label+'.jsonl')).open('w') as out:r=subprocess.run(args,stdout=out,stderr=subprocess.STDOUT)
 failures=[]
 for line in (ev/(label+'.jsonl')).read_text().splitlines():
  try:item=json.loads(line)
  except json.JSONDecodeError:continue
  if item.get('Action')=='fail' and item.get('Test'):failures.append(item['Test'])
 return {'exit':r.returncode,'failures':failures,'args':args}
pattern=(tmp/'pattern').read_text().replace('TestDemoFixturesLoad','TestEachBoardFixtureLoadsAtItsSpecSize')
changes={
 'internal/ask/ask.go':[
 ('func Create(ctx context.Context, home string, options CreateOptions) output.Envelope {','func Create(ctx context.Context, home string, options CreateOptions) output.Envelope {\n if true {return failed("validate", "feature removed", nil)}'),
 ('func AnswerText(record Record, replies []Reply, stopped bool) string {','func AnswerText(record Record, replies []Reply, stopped bool) string {\n if true {return ""}'),
 ('func DismissText(record Record) string {','func DismissText(record Record) string {\n if true {return ""}'),
 ('func Route(record Record, agents []records.Agent) (target records.Agent, stopped, ok bool) {','func Route(record Record, agents []records.Agent) (target records.Agent, stopped, ok bool) {\n if true {return records.Agent{}, true, false}')],
 'cmd/ask/ask.go':[('give the ask JSON with --file FILE or --file -','feature removed')],
 'internal/tui/tui.go':[('func (m *Model) View() tea.View {','func (m *Model) View() tea.View {\n if true {return tea.NewView("")}')],
 'internal/tui/ask.go':[
 ('func (m *Model) closeAsk(record ask.Record, action string, replies []ask.Reply) tea.Cmd {','func (m *Model) closeAsk(record ask.Record, action string, replies []ask.Reply) tea.Cmd {\n if true {return nil}'),
 ('func (m *Model) startAskDelivery() tea.Cmd {','func (m *Model) startAskDelivery() tea.Cmd {\n if true {return nil}')],
 'internal/tui/header.go':[('for _, record := range asks {','for _, record := range asks[:0] {')],
 'internal/tui/refresh.go':[('filepath.Join(state, "asks")','filepath.Join(state, "removed-asks")')],
 'internal/tui/component/chat.go':[('if m.Styled {','if false && m.Styled {')],
 'internal/boardexport/check.go':[(f'"{b}":',f'"Removed{b}":') for b in ['ChatAsk','ChatAskMulti','ChatAskPreview','ChatAskStopped','CompactChatAsk']],
}
results={};results['features-disabled']=run('features-disabled',changes,['./cmd/','./internal/ask/','./internal/boardexport/','./internal/tui/','./internal/tui/component/'],pattern)
results['blocked-caller-guard']=run('blocked-caller-guard',{'internal/tui/ask.go':[('if m.chatCallerBlocked {','if false && m.chatCallerBlocked {')]},['./internal/tui/'],'^TestABlockedCallerCannotAnswer$')
# The fixture regression guard must detect a broken fixture terminal size.
fixture=root/'.hive/spec/tui-v2/design/ChatAsk.fixture.json';data=json.loads(fixture.read_text());data['terminal']['width']=119;dest=tmp/'broken.fixture.json';dest.write_text(json.dumps(data));overlay=tmp/'fixture-overlay.json';overlay.write_text(json.dumps({'Replace':{str(fixture):str(dest)}}))
# os.ReadFile does not use go overlays; break the demo fixture loader instead.
results['fixture-guard']=run('fixture-guard',{'internal/tui/demo.go':[('func loadDemoFixture(path string) (*demoFixture, error) {','func loadDemoFixture(path string) (*demoFixture, error) {\n if true {return nil, fmt.Errorf("feature removed")}')]},['./internal/tui/'],'^TestEachBoardFixtureLoadsAtItsSpecSize$')
(ev/'mutations.json').write_text(json.dumps(results,indent=2)+'\n')
