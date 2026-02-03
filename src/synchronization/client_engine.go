package synchronization

import (
	"axial/api"
	"axial/models"
)

// ApplyIncomingMessages inserts messages from the server response into our store if missing.
func ApplyIncomingMessages(store api.SyncStore, messagesPeriods []models.MessagesPeriod) error {
	for _, mp := range messagesPeriods {
		ourMessages, err := store.GetMessagesByPeriod(mp.Period)
		if err != nil {
			return err
		}
		for _, msg := range mp.Messages {
			if !msg.In(ourMessages) {
				if err := store.UpsertMessage(msg); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// CollectMissingInRemote computes local messages that were not present in the server's payload.
func CollectMissingInRemote(store api.SyncStore, messagesPeriods []models.MessagesPeriod) ([]models.Message, error) {
	missing := []models.Message{}
	for _, mp := range messagesPeriods {
		ourMessages, err := store.GetMessagesByPeriod(mp.Period)
		if err != nil {
			return nil, err
		}
		for _, local := range ourMessages {
			if !local.In(mp.Messages) {
				missing = append(missing, local)
			}
		}
	}
	return missing, nil
}

// NextRequestFromHashes builds the next drill-down request based on hashed time ranges and user ranges.
func NextRequestFromHashes(store api.SyncStore, response api.SyncResponse) (periods []models.HashedPeriod, users []models.HashedUsersRange, err error) {
	periods = []models.HashedPeriod{}
	users = []models.HashedUsersRange{}

	// For each hashed period from the server, compare with our hash; include mismatches.
	if len(response.Ranges) > 0 {
		periodsToHash := make([]models.Period, 0, len(response.Ranges))
		for _, hp := range response.Ranges {
			periodsToHash = append(periodsToHash, hp.Period)
		}
		ourHashes, err := store.GenerateHashRanges(periodsToHash)
		if err != nil {
			return nil, nil, err
		}
		for _, our := range ourHashes {
			start := models.RealizeStart(our.Start)
			end := models.RealizeEnd(our.End)
			for _, theirs := range response.Ranges {
				if models.RealizeStart(theirs.Start) == start && models.RealizeEnd(theirs.End) == end && our.Hash != theirs.Hash {
					periods = append(periods, theirs)
				}
			}
		}
	}

	// For each user hashed range from the server, compare with our hash; include mismatches.
	for _, ur := range response.UserRangeHashes {
		ourHash, err := store.GetUsersHashByFingerprintRange(ur.Start, ur.End)
		if err != nil {
			return nil, nil, err
		}
		if ourHash != ur.Hash {
			users = append(users, ur)
		}
	}

	return periods, users, nil
}
