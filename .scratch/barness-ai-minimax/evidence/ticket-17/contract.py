"""Finite observed historical stages, with fixture failures kept distinct from REDs."""
SOURCE='230691742b731c77e8167b94646219e10f78c8f7'
ARCHIVE='.evidence/minimax-retained/ticket-17-final-payload/offline-tdd-raw-v2.tar.gz'
# Fixed before retention code: original commands/cases, not inferred success.
STAGES={
 'red-containers':(1,{}),
 'red-containers-actual':(1,{'FAIL':5}),
 'green-containers':(0,{'PASS':5}),
 'red-message-fields':(1,{'FAIL':9}),
 'green-message-fields':(0,{'PASS':14}),
 'red-block-fields':(1,{'FAIL':14}),
 'green-block-fields':(0,{'PASS':28}),
 'red-tool-fields':(1,{'FAIL':14}),
 'red-tools-with-business-controls':(1,{'FAIL':15,'PASS':1}),
 'green-final-shapes':(1,{'PASS':29,'FAIL':15}),
 'red-tools-native-containers-actual':(1,{'FAIL':14,'PASS':2}),
 'green-final-native-shapes':(0,{'PASS':44}),
 'red-expanded-trace-old-constant':(1,{'FAIL':2,'PASS':4}),
 'green-expanded-trace':(0,{'PASS':6}),
 'regression-ordinary':(1,{'PASS':1662,'NOT_RUN':8,'FAIL':1}),
 'diagnose-catalog-before-local-dependencies':(1,{'FAIL':2}),
 'vet':(0,{}),'vet-live':(0,{}),
 'regression-ordinary-restored':(0,{'PASS':1663,'NOT_RUN':8}),
 'regression-race-restored':(0,{'PASS':1663,'NOT_RUN':8}),
 'selected-opt-in-ordinary':(0,{'PASS':8}),'selected-opt-in-race':(0,{'PASS':8}),
 'vet-restored':(0,{}),'vet-live-restored':(0,{})}
TRUE_RED={'red-containers-actual':5,'red-message-fields':9,'red-block-fields':14,'red-tools-native-containers-actual':14}
FINAL=('regression-ordinary-restored','regression-race-restored','selected-opt-in-ordinary','selected-opt-in-race','vet-restored','vet-live-restored')
CHANGED=('ai/anthropic_minimax_payload.go','ai/anthropic_minimax_message.go','ai/e2e/minimax_messages_final_shape_test.go','ai/e2e/minimax_messages_final_tools_test.go','ai/release/traceability.json','ai/internal/testkit/release/plan.go')
NOT_RUN={*[f'MM05-design-load-{host}-{route}' for host in ('local','cloud') for route in ('chat','responses','messages')],'MM06-frozen-catalog-metadata','PIDIFF-MM06-catalog-differences'}
