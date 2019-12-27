package powerdns

import (
	"gopkg.in/jarcoal/httpmock.v1"
	"net/http"
	"strings"
	"testing"
)

func TestListZones(t *testing.T) {
	testDomain := generateTestZone(true)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", generateTestAPIVHostURL()+"/zones",
		func(req *http.Request) (*http.Response, error) {
			if req.Header.Get("X-Api-Key") == testAPIKey {
				zonesMock := []Zone{
					{
						ID:             String(fixDomainSuffix(testDomain)),
						Name:           String(fixDomainSuffix(testDomain)),
						URL:            String("/api/v1/servers/" + testVHost + "/zones/" + fixDomainSuffix(testDomain)),
						Kind:           ZoneKindPtr(NativeZoneKind),
						Serial:         Uint32(1337),
						NotifiedSerial: Uint32(1337),
					},
				}
				return httpmock.NewJsonResponse(http.StatusOK, zonesMock)
			}
			return httpmock.NewStringResponse(http.StatusUnauthorized, "Unauthorized"), nil
		},
	)

	p := initialisePowerDNSTestClient()
	zones, err := p.Zones.List()
	if err != nil {
		t.Errorf("%s", err)
	}
	if len(zones) == 0 {
		t.Error("Received amount of statistics is 0")
	}
}

func TestListZonesError(t *testing.T) {
	p := initialisePowerDNSTestClient()
	p.Hostname = "invalid"
	if _, err := p.Zones.List(); err == nil {
		t.Error("error is nil")
	}
}

func TestGetZone(t *testing.T) {
	testDomain := generateTestZone(true)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", generateTestAPIVHostURL()+"/zones/"+testDomain,
		func(req *http.Request) (*http.Response, error) {
			if req.Header.Get("X-Api-Key") == testAPIKey {
				zoneMock := Zone{
					ID:   String(fixDomainSuffix(testDomain)),
					Name: String(fixDomainSuffix(testDomain)),
					URL:  String("/api/v1/servers/" + testVHost + "/zones/" + fixDomainSuffix(testDomain)),
					Kind: ZoneKindPtr(NativeZoneKind),
					RRsets: []RRset{
						{
							Name: String(fixDomainSuffix(testDomain)),
							Type: String("SOA"),
							TTL:  Uint32(3600),
							Records: []Record{
								{
									Content: String("a.misconfigured.powerdns.server. hostmaster." + fixDomainSuffix(testDomain) + " 1337 10800 3600 604800 3600"),
								},
							},
						},
					},
					Serial:         Uint32(1337),
					NotifiedSerial: Uint32(1337),
				}
				return httpmock.NewJsonResponse(http.StatusOK, zoneMock)
			}
			return httpmock.NewStringResponse(http.StatusUnauthorized, "Unauthorized"), nil
		},
	)

	p := initialisePowerDNSTestClient()
	zone, err := p.Zones.Get(testDomain)
	if err != nil {
		t.Errorf("%s", err)
	}
	if *zone.ID != fixDomainSuffix(testDomain) {
		t.Error("Received no zone")
	}
}

func TestGetZonesError(t *testing.T) {
	testDomain := generateTestZone(false)
	p := initialisePowerDNSTestClient()
	p.Hostname = "invalid"
	if _, err := p.Zones.Get(testDomain); err == nil {
		t.Error("error is nil")
	}
}

func TestAddNativeZone(t *testing.T) {
	testDomain := generateTestZone(false)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", generateTestAPIVHostURL()+"/zones",
		func(req *http.Request) (*http.Response, error) {
			if req.Header.Get("X-Api-Key") == testAPIKey {
				zoneMock := Zone{
					ID:   String(fixDomainSuffix(testDomain)),
					Name: String(fixDomainSuffix(testDomain)),
					Type: ZoneTypePtr(ZoneZoneType),
					URL:  String("api/v1/servers/" + testVHost + "/zones/" + fixDomainSuffix(testDomain)),
					Kind: ZoneKindPtr(NativeZoneKind),
					RRsets: []RRset{
						{
							Name: String(fixDomainSuffix(testDomain)),
							Type: String("SOA"),
							TTL:  Uint32(3600),
							Records: []Record{
								{
									Content:  String("a.misconfigured.powerdns.server. hostmaster." + fixDomainSuffix(testDomain) + " 0 10800 3600 604800 3600"),
									Disabled: Bool(false),
								},
							},
						},
						{
							Name: String(fixDomainSuffix(testDomain)),
							Type: String("NS"),
							TTL:  Uint32(3600),
							Records: []Record{
								{
									Content:  String("ns.example.tld."),
									Disabled: Bool(false),
								},
							},
						},
					},
					Serial:      Uint32(0),
					Masters:     []string{},
					DNSsec:      Bool(true),
					Nsec3Param:  String(""),
					Nsec3Narrow: Bool(false),
					SOAEdit:     String("foo"),
					SOAEditAPI:  String("foo"),
					APIRectify:  Bool(true),
					Account:     String(""),
				}
				return httpmock.NewJsonResponse(http.StatusCreated, zoneMock)
			}
			return httpmock.NewStringResponse(http.StatusUnauthorized, "Unauthorized"), nil
		},
	)

	p := initialisePowerDNSTestClient()
	zone, err := p.Zones.AddNative(testDomain, true, "", false, "foo", "foo", true, []string{"ns.foo.tld."})
	if err != nil {
		t.Errorf("%s", err)
	}
	if *zone.ID != fixDomainSuffix(testDomain) || *zone.Kind != NativeZoneKind {
		t.Error("Zone wasn't created")
	}
}

func TestAddNativeZoneError(t *testing.T) {
	testDomain := generateTestZone(false)
	p := initialisePowerDNSTestClient()
	p.Hostname = "invalid"
	if _, err := p.Zones.AddNative(testDomain, true, "", false, "foo", "foo", true, []string{"ns.foo.tld."}); err == nil {
		t.Error("error is nil")
	}
}

func TestAddMasterZone(t *testing.T) {
	testDomain := generateTestZone(false)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", generateTestAPIVHostURL()+"/zones",
		func(req *http.Request) (*http.Response, error) {
			if req.Header.Get("X-Api-Key") == testAPIKey {
				zoneMock := Zone{
					ID:   String(fixDomainSuffix(testDomain)),
					Name: String(fixDomainSuffix(testDomain)),
					Type: ZoneTypePtr(ZoneZoneType),
					URL:  String("api/v1/servers/" + testVHost + "/zones/" + fixDomainSuffix(testDomain)),
					Kind: ZoneKindPtr(MasterZoneKind),
					RRsets: []RRset{
						{
							Name: String(fixDomainSuffix(testDomain)),
							Type: String("SOA"),
							TTL:  Uint32(3600),
							Records: []Record{
								{
									Content:  String("a.misconfigured.powerdns.server. hostmaster." + fixDomainSuffix(testDomain) + " 0 10800 3600 604800 3600"),
									Disabled: Bool(false),
								},
							},
						},
						{
							Name: String(fixDomainSuffix(testDomain)),
							Type: String("NS"),
							TTL:  Uint32(3600),
							Records: []Record{
								{
									Content:  String("ns.example.tld."),
									Disabled: Bool(false),
								},
							},
						},
					},
					Serial:      Uint32(0),
					Masters:     []string{},
					DNSsec:      Bool(true),
					Nsec3Param:  String(""),
					Nsec3Narrow: Bool(false),
					SOAEdit:     String("foo"),
					SOAEditAPI:  String("foo"),
					APIRectify:  Bool(true),
					Account:     String(""),
				}
				return httpmock.NewJsonResponse(http.StatusCreated, zoneMock)
			}
			return httpmock.NewStringResponse(http.StatusUnauthorized, "Unauthorized"), nil
		},
	)

	p := initialisePowerDNSTestClient()
	zone, err := p.Zones.AddMaster(testDomain, true, "", false, "foo", "foo", true, []string{"ns.foo.tld."})
	if err != nil {
		t.Errorf("%s", err)
	}
	if *zone.ID != fixDomainSuffix(testDomain) || *zone.Kind != MasterZoneKind {
		t.Error("Zone wasn't created")
	}
}

func TestAddMasterZoneError(t *testing.T) {
	testDomain := generateTestZone(false)
	p := initialisePowerDNSTestClient()
	p.Hostname = "invalid"
	if _, err := p.Zones.AddMaster(testDomain, true, "", false, "foo", "foo", true, []string{"ns.foo.tld."}); err == nil {
		t.Error("error is nil")
	}
}

func TestAddSlaveZone(t *testing.T) {
	testDomain := generateTestZone(false)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", generateTestAPIVHostURL()+"/zones",
		func(req *http.Request) (*http.Response, error) {
			if req.Header.Get("X-Api-Key") == testAPIKey {
				zoneMock := Zone{
					ID:          String(fixDomainSuffix(testDomain)),
					Name:        String(fixDomainSuffix(testDomain)),
					Type:        ZoneTypePtr(ZoneZoneType),
					URL:         String("api/v1/servers/" + testVHost + "/zones/" + fixDomainSuffix(testDomain)),
					Kind:        ZoneKindPtr(SlaveZoneKind),
					Serial:      Uint32(0),
					Masters:     []string{"127.0.0.1"},
					DNSsec:      Bool(true),
					Nsec3Param:  String(""),
					Nsec3Narrow: Bool(false),
					SOAEdit:     String(""),
					SOAEditAPI:  String("DEFAULT"),
					APIRectify:  Bool(true),
					Account:     String(""),
				}
				return httpmock.NewJsonResponse(http.StatusCreated, zoneMock)
			}
			return httpmock.NewStringResponse(http.StatusUnauthorized, "Unauthorized"), nil
		},
	)

	p := initialisePowerDNSTestClient()
	zone, err := p.Zones.AddSlave(testDomain, []string{"127.0.0.1"})
	if err != nil {
		t.Errorf("%s", err)
	}
	if *zone.ID != fixDomainSuffix(testDomain) || *zone.Kind != SlaveZoneKind {
		t.Error("Zone wasn't created")
	}
}

func TestAddSlaveZoneError(t *testing.T) {
	testDomain := generateTestZone(false)
	p := initialisePowerDNSTestClient()
	p.Hostname = "invalid"
	if _, err := p.Zones.AddSlave(testDomain, []string{"ns5.foo.tld."}); err == nil {
		t.Error("error is nil")
	}
}

func TestChangeZone(t *testing.T) {
	testDomain := generateTestZone(true)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("PUT", generateTestAPIVHostURL()+"/zones/"+testDomain,
		func(req *http.Request) (*http.Response, error) {
			if req.Header.Get("X-Api-Key") == testAPIKey {
				return httpmock.NewBytesResponse(http.StatusNoContent, []byte{}), nil
			}
			return httpmock.NewStringResponse(http.StatusUnauthorized, "Unauthorized"), nil
		},
	)

	p := initialisePowerDNSTestClient()

	t.Run("ChangeValidZone", func(t *testing.T) {
		if err := p.Zones.Change(testDomain, &Zone{Nameservers: []string{"ns23.foo.tld."}}); err != nil {
			t.Errorf("%s", err)
		}
	})
	t.Run("ChangeInvalidZone", func(t *testing.T) {
		if err := p.Zones.Change("doesnt-exist", &Zone{Nameservers: []string{"ns23.foo.tld."}}); err == nil {
			t.Errorf("Changing an invalid zone does not return an error")
		}
	})
}

func TestChangeZoneError(t *testing.T) {
	testDomain := generateTestZone(false)
	p := initialisePowerDNSTestClient()
	p.Hostname = "invalid"
	if err := p.Zones.Change(testDomain, &Zone{Nameservers: []string{"ns23.foo.tld."}}); err == nil {
		t.Error("error is nil")
	}
}

func TestDeleteZone(t *testing.T) {
	testDomain := generateTestZone(true)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("DELETE", generateTestAPIVHostURL()+"/zones/"+testDomain,
		func(req *http.Request) (*http.Response, error) {
			if req.Header.Get("X-Api-Key") == testAPIKey {
				return httpmock.NewBytesResponse(http.StatusNoContent, []byte{}), nil
			}
			return httpmock.NewStringResponse(http.StatusUnauthorized, "Unauthorized"), nil
		},
	)

	p := initialisePowerDNSTestClient()
	if err := p.Zones.Delete(testDomain); err != nil {
		t.Errorf("%s", err)
	}
}

func TestDeleteZoneError(t *testing.T) {
	testDomain := generateTestZone(false)
	p := initialisePowerDNSTestClient()
	p.Hostname = "invalid"
	if err := p.Zones.Delete(testDomain); err == nil {
		t.Error("error is nil")
	}
}

func TestNotify(t *testing.T) {
	testDomain := generateTestZone(true)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	registerZoneMockResponder(testDomain)
	httpmock.RegisterResponder("PUT", generateTestAPIVHostURL()+"/zones/"+testDomain+"/notify",
		func(req *http.Request) (*http.Response, error) {
			if req.Header.Get("X-Api-Key") == testAPIKey {
				return httpmock.NewStringResponse(http.StatusOK, "{\"result\":\"Notification queued\"}"), nil
			}
			return httpmock.NewStringResponse(http.StatusUnauthorized, "Unauthorized"), nil
		},
	)

	p := initialisePowerDNSTestClient()
	notifyResult, err := p.Zones.Notify(testDomain)
	if err != nil {
		t.Errorf("%s", err)
	}
	if *notifyResult.Result != "Notification queued" {
		t.Error("Notification was not queued successfully")
	}
}

func TestNotifyError(t *testing.T) {
	testDomain := generateTestZone(false)
	p := initialisePowerDNSTestClient()
	p.Hostname = "invalid"
	if _, err := p.Zones.Notify(testDomain); err == nil {
		t.Error("error is nil")
	}
}

func TestExport(t *testing.T) {
	testDomain := generateTestZone(true)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	registerZoneMockResponder(testDomain)
	httpmock.RegisterResponder("GET", generateTestAPIVHostURL()+"/zones/"+testDomain+"/export",
		func(req *http.Request) (*http.Response, error) {
			if req.Header.Get("X-Api-Key") == testAPIKey {
				return httpmock.NewStringResponse(http.StatusOK, fixDomainSuffix(testDomain)+"	3600	SOA	a.misconfigured.powerdns.server. hostmaster."+fixDomainSuffix(testDomain)+" 1 10800 3600 604800 3600"), nil
			}
			return httpmock.NewStringResponse(http.StatusUnauthorized, "Unauthorized"), nil
		},
	)

	p := initialisePowerDNSTestClient()
	export, err := p.Zones.Export(testDomain)
	if err != nil {
		t.Errorf("%s", err)
	}
	if !strings.HasPrefix(string(export), testDomain) {
		t.Errorf("Export payload wrong: \"%s\"", export)
	}
}

func TestExportError(t *testing.T) {
	testDomain := generateTestZone(false)
	p := initialisePowerDNSTestClient()
	p.Hostname = "invalid"
	if _, err := p.Zones.Export(testDomain); err == nil {
		t.Error("error is nil")
	}
}
