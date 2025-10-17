const pad = (v, n = 2) => {
    v = v + ""; // convert to string
    if (v.length >= n)
        return v;
    for (let i = 0; i < n; i++) {
        v = "0" + v;
        if (v.length >= n)
            break;
    }
    return v;
};

const DATE_TYPE = { INPUT: 0, DISPLAY_FULL: 1, DISPLAY_DATE: 2 };

const formatDate = (d, type) => {
    if (typeof d === "string") {
        d = d.trim();
    }

    if (!d)
        return "";

    if (typeof d === "string" || typeof d === "number") {
        d = new Date(d);
    } else if (!(d instanceof Date)) {
        return "";
    }

    if (isNaN(d.getTime()))
        return "";

    const date = d.getFullYear() + "-" + pad(d.getMonth() + 1) + "-" + pad(d.getDate());
    const time = pad(d.getHours()) + ":" + pad(d.getMinutes());

    if (type === DATE_TYPE.INPUT)
        return date + "T" + time; // YYYY-MM-DDThh:mm - 2022-01-07T23:43
    else if (type === DATE_TYPE.DISPLAY_FULL)
        return date + " " + time; // YYYY-MM-DD hh:mm - 2022-01-07 23:43
    else if (type === DATE_TYPE.DISPLAY_DATE)
        return date; // YYYY-MM-DD - 2022-01-07
    else return date + " " + time; // YYYY-MM-DD hh:mm - 2022-01-07 23:43
};

// TODO: write unit tests
const getDuration = (start, end) => {
    if (!start || !end) return "";

    const startTime = new Date(start);
    const endTime = new Date(end);

    const diffMs = endTime - startTime;
    const diffMins = Math.floor(diffMs / 60 / 1000);
    const diffSecs = Math.floor((diffMs % (60 * 1000)) / 1000);
    return `${diffMins}m ${diffSecs}s`;
};

function getRelativeTime(timestamp) {
    const now = Date.now();
    const diffMs = now - timestamp;
    const absDiff = Math.abs(diffMs);

    // Define time units in milliseconds
    const units = [
        { name: "year", ms: 365.25 * 24 * 60 * 60 * 1000 },
        { name: "month", ms: 30.44 * 24 * 60 * 60 * 1000 },
        { name: "week", ms: 7 * 24 * 60 * 60 * 1000 },
        { name: "day", ms: 24 * 60 * 60 * 1000 },
        { name: "hour", ms: 60 * 60 * 1000 },
        { name: "minute", ms: 60 * 1000 },
        { name: "second", ms: 1000 }
    ];

    // Find the appropriate unit
    for (const unit of units) {
        const value = Math.floor(absDiff / unit.ms);
        if (value >= 1) {
            const plural = value > 1 ? "s" : "";
            const preposition = diffMs < 0 ? "in " : "";
            const suffix = diffMs < 0 ? "" : " ago";
            return `${preposition}${value} ${unit.name}${plural}${suffix}`;
        }
    }

    return "just now";
}

export { pad, formatDate, getDuration, DATE_TYPE, getRelativeTime };