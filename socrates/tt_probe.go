package socrates

func (r *RuleEngine) ttProbe(hash uint64) (ttEntry, bool) {
	entry := r.tt[hash&TTMask]
	if entry.hash == hash {
		return entry, true
	}
	return ttEntry{}, false
}
