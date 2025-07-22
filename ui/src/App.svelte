<script lang="ts">
    import {onMount} from 'svelte';
    import {Control, LatLng, type LatLngExpression, Map as LeafletMap, Polyline, TileLayer} from 'leaflet';
    import {AISTrackSymbol, type PositionReport, type ShipStaticData} from '@arl/leaflet-tracksymbol2';

    console.log(window.nmeaLogger);

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

        for (const [id, positionReportEntries] of Object.entries(window.nmeaLogger.positionReportsMap)) {
            const latLngs: LatLngExpression[] = [];
            for (const positionReportEntry of positionReportEntries) {
                latLngs.push(new LatLng(positionReportEntry.positionReport.Latitude, positionReportEntry.positionReport.Longitude));
            }
            const path = new Polyline(latLngs).addTo(map);
            path.bindTooltip(id);

            const lastPositionReport = positionReportEntries[positionReportEntries.length - 1].positionReport;
            const positionReport: PositionReport = {
                navigationalStatus: lastPositionReport.NavigationalStatus,
                rateOfTurn: lastPositionReport.RateOfTurn,
                sog: lastPositionReport.Sog,
                positionAccuracy: lastPositionReport.PositionAccuracy,
                longitude: lastPositionReport.Longitude,
                latitude: lastPositionReport.Latitude,
                cog: lastPositionReport.Cog,
                trueHeading: lastPositionReport.TrueHeading,
            };
            const shipStaticData0 = window.nmeaLogger.shipStaticDataMap[id];
            let shipStaticData: ShipStaticData | undefined = undefined;
            if (shipStaticData0 !== undefined) {
                shipStaticData = {
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
            }
            new AISTrackSymbol(positionReport, {
                shipStaticData: shipStaticData,
            }).addTo(map);
        }
    });
</script>

<div bind:this={mapElement} id="map">
</div>

<style>
</style>
