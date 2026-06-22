package scaner

import (
	"errors"
	"testing"
)

type fakeRow struct {
	err error
}

func (r fakeRow) Scan(dest ...any) error { return r.err }

func TestScannersSuccess(t *testing.T) {
	row := fakeRow{}
	calls := []func() error{
		func() error { _, err := ScanCallAnalysis(row); return err },
		func() error { _, err := ScanAnalysisInstruction(row); return err },
		func() error { _, err := ScanCall(row); return err },
		func() error { _, err := ScanCompany(row); return err },
		func() error { _, err := ScanCompanyMember(row); return err },
		func() error { _, err := ScanDepartment(row); return err },
		func() error { _, err := ScanDepartmentMember(row); return err },
		func() error { _, err := ScanInvitation(row); return err },
		func() error { _, err := ScanProcessingJob(row); return err },
		func() error { _, err := ScanRefreshSession(row); return err },
		func() error { _, err := ScanTranscription(row); return err },
		func() error { _, err := ScanUser(row); return err },
	}
	for i, call := range calls {
		if err := call(); err != nil {
			t.Fatalf("scanner %d: %v", i, err)
		}
	}
}

func TestScannersPropagateErrors(t *testing.T) {
	want := errors.New("scan failed")
	row := fakeRow{err: want}
	calls := []func() error{
		func() error { _, err := ScanCallAnalysis(row); return err },
		func() error { _, err := ScanAnalysisInstruction(row); return err },
		func() error { _, err := ScanCall(row); return err },
		func() error { _, err := ScanCompany(row); return err },
		func() error { _, err := ScanCompanyMember(row); return err },
		func() error { _, err := ScanDepartment(row); return err },
		func() error { _, err := ScanDepartmentMember(row); return err },
		func() error { _, err := ScanInvitation(row); return err },
		func() error { _, err := ScanProcessingJob(row); return err },
		func() error { _, err := ScanRefreshSession(row); return err },
		func() error { _, err := ScanTranscription(row); return err },
		func() error { _, err := ScanUser(row); return err },
	}
	for i, call := range calls {
		if err := call(); !errors.Is(err, want) {
			t.Fatalf("scanner %d error = %v", i, err)
		}
	}
}
