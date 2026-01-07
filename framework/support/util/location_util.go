package util

import "strings"

type Location struct {
	WhsID     string
	ZoneID    string
	RackID    string
	FloorID   string
	GridID    string
	PathwayID string
}

func (l Location) GetZoneId() string {
	return l.WhsID + "-" + l.ZoneID
}

func (l Location) GetPathwayId() string {
	return l.WhsID + "-" + l.ZoneID + "-" + l.PathwayID
}

func (l Location) GetSegmentId() string {
	return l.WhsID + "-" + l.ZoneID + "-" + l.PathwayID + "-" + l.RackID
}

func (l Location) GetFloorId() string {
	return l.GetSegmentId() + "-" + l.FloorID
}

func (l Location) GetGridId() string {
	return l.GetFloorId() + "-" + l.GridID
}

func UnmarshalLocationID(locationID string) *Location {
	location := &Location{}
	segmentList := strings.Split(locationID, "-")

	//兼容老的location(5位，无 pathway);
	if len(segmentList) == 5 {
		location.WhsID = segmentList[0]
		location.ZoneID = segmentList[1]
		location.RackID = segmentList[2]
		location.FloorID = segmentList[3]
		location.GridID = segmentList[4]
	}

	if len(segmentList) == 6 {
		location.WhsID = segmentList[0]
		location.ZoneID = segmentList[1]
		location.PathwayID = segmentList[2]
		location.RackID = segmentList[3]
		location.FloorID = segmentList[4]
		location.GridID = segmentList[5]
	}

	return location
}
