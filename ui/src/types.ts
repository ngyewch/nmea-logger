export interface NMEALoggerData {
    shipStaticDataMap: Record<string, ShipStaticData>;
    positionReportsMap: Record<string, PositionReportEntry[]>;
}

export interface ShipStaticData {
    ImoNumber: number;
    CallSign: string;
    Name: string;
    Type: number;
    Dimension: Dimension;
    FixType: number;
    Eta: ETA;
    MaximumStaticDraught: number;
    Destination: string;
    Dte: boolean;
    Spare: boolean;
}

export interface Dimension {
    A: number;
    B: number;
    C: number;
    D: number;
}

export interface ETA {
    Month: number;
    Day: number;
    Hour: number;
    Minute: number;
}

export interface PositionReportEntry {
    t: number;
    positionReport: PositionReport;
}

export interface PositionReport {
    NavigationalStatus: number;
    RateOfTurn: number;
    Sog: number;
    PositionAccuracy: boolean;
    Longitude: number;
    Latitude: number;
    Cog: number;
    TrueHeading: number;
    Timestamp: number;
    SpecialManoeuvreIndicator: number;
    Spare: number;
    Raim: boolean;
}