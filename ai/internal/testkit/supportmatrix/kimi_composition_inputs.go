package supportmatrix

import (
	"path/filepath"
)

// KimiCompositionInputs closes over both complete native inventories, producer
// process evidence/executables and trusted reservations. No referenced material
// is copied into, or silently omitted from, the flat descriptor directory.
func KimiCompositionInputs(dir string) (map[string]string, error) {
	q, e := VerifyKimiComposition(dir, false)
	if e != nil {
		return nil, e
	}
	root, e := kimiRepositoryRoot()
	if e != nil {
		return nil, e
	}
	raw, e := regionalReadFile(filepath.Join(dir, KimiCompositionFile))
	if e != nil {
		return nil, e
	}
	inputs := map[string]string{KimiCompositionFile: glmHash(raw)}
	for _, m := range q.Measurements {
		native, e := kimiLocalPath(root, m.Evidence.Bundle)
		if e != nil {
			return nil, e
		}
		files, e := regionalInventory(native)
		if e != nil {
			return nil, e
		}
		seal, e := regionalReadFile(filepath.Join(native, regionalSealFile))
		if e != nil {
			return nil, e
		}
		files[regionalSealFile] = glmHash(seal)
		for name, h := range files {
			inputs["@kimi/"+filepath.ToSlash(filepath.Join(m.Evidence.Bundle, name))] = h
		}
		producer, e := verifyKimiProducer(root, native, m.Evidence.Producer, m.Original)
		if e != nil {
			return nil, e
		}
		for name, h := range producer {
			inputs["@kimi/"+name] = h
		}
		var g RegionalGrant
		if readJSON(filepath.Join(native, "regional-grant.json"), &g) != nil {
			return nil, refuse("Kimi original grant input missing")
		}
		store, e := regionalReservationStore(g.Combo)
		if e != nil {
			return nil, e
		}
		id, scope := regionalReservationKeys(g)
		for _, name := range []string{id, scope} {
			raw, e := regionalReadFile(filepath.Join(store, name))
			if e != nil {
				return nil, e
			}
			// The store is selected by trusted Git authority, never a descriptor path.
			inputs["@kimi-authority/"+g.Combo+"/"+name] = glmHash(raw)
		}
	}
	return inputs, nil
}
