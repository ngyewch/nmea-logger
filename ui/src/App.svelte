<script lang="ts">
    import {onMount} from 'svelte';
    import {Control, Map as LeafletMap, TileLayer} from 'leaflet';
    import {AISTrackSymbol, type PositionReport, type ShipStaticData} from '@arl/leaflet-tracksymbol2';
    import {type AISRecord} from './types.js';
    import {CustomControl} from './customControl.js';

    const openStreetMapTileLayer = new TileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        minZoom: 0,
        maxZoom: 20,
        maxNativeZoom: 19,
        attribution: "© OpenStreetMap contributors",
    });
    const openSeaMapTileLayer = new TileLayer('https://tiles.openseamap.org/seamark/{z}/{x}/{y}.png', {
        opacity: 0.2,
        attribution: 'Map data: &copy; <a href="http://www.openseamap.org">OpenSeaMap</a> contributors',
    });

    let mapElement: HTMLDivElement;
    let map: LeafletMap;

    function resolveWsUrl(path: string): string {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const host = window.location.host;
        const p = window.location.pathname.lastIndexOf('/');
        const basePath = window.location.pathname.substring(0, p + 1);
        const resolvedPath = path.startsWith('/') ? path : basePath + path;
        return `${protocol}//${host}${resolvedPath}`;
    }

    const wsUrl = resolveWsUrl('/ws');

    let trackSymbolMap: Record<number, AISTrackSymbol> = {};

    function getTrackSymbol(userId: number, positionReport: PositionReport | undefined, shipStaticData: ShipStaticData | undefined): AISTrackSymbol | undefined {
        if (map === undefined) {
            return undefined;
        }
        let trackSymbol = trackSymbolMap[userId];
        if (trackSymbol === undefined) {
            if (positionReport === undefined) {
                return undefined;
            }
            trackSymbol = new AISTrackSymbol(positionReport, {
                shipStaticData: shipStaticData,
            });
            trackSymbol.addTo(map);
            trackSymbolMap[userId] = trackSymbol;
        }
        return trackSymbol;
    }

    onMount(() => {
        if (!mapElement) {
            return;
        }
        map = new LeafletMap(mapElement, {
            center: [1.251, 103.826],
            zoom: 13,
        });
        openStreetMapTileLayer.addTo(map);

        const baseMaps = {
            "OpenStreetMaps": openStreetMapTileLayer,
        };
        const overlayMaps = {
            "OpenSeaMap": openSeaMapTileLayer,
        };
        const layersOptions = {};
        new Control.Layers(baseMaps, overlayMaps, layersOptions).addTo(map);

        const scaleOptions = {
            maxWidth: 200,
        };
        new Control.Scale(scaleOptions).addTo(map);

        const customControl = new CustomControl({
            position: 'bottomright',
        });
        customControl.addTo(map);
        let t: number = 0;

        const ws = new WebSocket(wsUrl);
        ws.onopen = () => {
            console.log("WebSocket connection opened.");
        };
        ws.onmessage = (event) => {
            const record = JSON.parse(event.data) as AISRecord;
            if (record.t > t) {
                customControl.setText(new Date(t).toISOString());
                t = record.t;
            }
            switch (record.type) {
                case 'positionReport':
                    const positionReport0 = record.positionReport;
                    const positionReport: PositionReport = {
                        userId: positionReport0.UserID,
                        navigationalStatus: positionReport0.NavigationalStatus,
                        rateOfTurn: positionReport0.RateOfTurn,
                        sog: positionReport0.Sog,
                        positionAccuracy: positionReport0.PositionAccuracy,
                        longitude: positionReport0.Longitude,
                        latitude: positionReport0.Latitude,
                        cog: positionReport0.Cog,
                        trueHeading: positionReport0.TrueHeading,
                    };
                    try {
                        const trackSymbol = getTrackSymbol(positionReport0.UserID, positionReport, undefined);
                        if (trackSymbol !== undefined) {
                            trackSymbol.setPositionReport(positionReport);
                        }
                    } catch (e) {
                        console.error(e);
                    }
                    break;

                case 'shipStaticData':
                    const shipStaticData0 = record.shipStaticData;
                    const shipStaticData: ShipStaticData = {
                        userId: shipStaticData0.UserID,
                        imoNumber: shipStaticData0.ImoNumber,
                        callSign: shipStaticData0.CallSign,
                        name: shipStaticData0.Name,
                        type: shipStaticData0.Type,
                        dimension: {
                            A: shipStaticData0.Dimension.A,
                            B: shipStaticData0.Dimension.B,
                            C: shipStaticData0.Dimension.C,
                            D: shipStaticData0.Dimension.D,
                        },
                        fixType: shipStaticData0.FixType,
                        eta: {
                            month: shipStaticData0.Eta.Month,
                            day: shipStaticData0.Eta.Day,
                            hour: shipStaticData0.Eta.Hour,
                            minute: shipStaticData0.Eta.Minute,
                        },
                        maximumStaticDraught: shipStaticData0.MaximumStaticDraught,
                        destination: shipStaticData0.Destination,
                        dte: shipStaticData0.Dte,
                    };
                    try {
                        const trackSymbol = getTrackSymbol(shipStaticData0.UserID, undefined, shipStaticData);
                        if (trackSymbol !== undefined) {
                            trackSymbol.setShipStaticData(shipStaticData);
                        }
                    } catch (e) {
                        console.error(e);
                    }
                    break;
            }
        };
        ws.onerror = (error) => {
            console.error("WebSocket error:", error);
        };
        ws.onclose = () => {
            console.log("WebSocket connection closed.");
        };
    });
</script>

<div bind:this={mapElement} id="map">
</div>

<style>
</style>
