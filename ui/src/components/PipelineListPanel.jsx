import { useState, useEffect, useRef } from "react";
import { request } from "../utils/Fetch";

import { useAppContext } from "../AppContext";
import { NotificationType, useNotificationContext } from "./Notification";

const PipelineListPanel = () => {
    const { display: displayNotification } = useNotificationContext();
    const { pipelines, setPipelines, setSelectedPipeline, setShowDetails } = useAppContext();
    const [searchFilter, setSearchFilter] = useState("");

    const pipelinesContainer = useRef();

    useEffect(() => {
        request({ url: "/api/pipelines",
            callback: ({ msg, success, json }) => {
                if (success) {
                    setPipelines(json);
                } else {
                    displayNotification({ message: "An error occurred fetching pipeline runs: " + msg, type: NotificationType.Error });
                }
            }
        });
    }, []);

    const showDetails = (name) => {
        setSelectedPipeline(name);
        setShowDetails(true);
    };

    const streamSorter = (a, b) => {
        // TODO: update with sorting priorities
        return a.name?.localeCompare(b.name);
    };

    const streamFilter = ({ name }) => searchFilter === "" || name.toLowerCase().includes(searchFilter.toLowerCase());

    return (
        <div className="w-full lg:w-1/3 h-screen overflow-y-hidden bg-indigo-900 flex flex-col">
            {/* add form */}

            {/* Search filter */}
            <div className="relative flex rounded-md shadow-sm">
                <input type="text" placeholder="search filter" value={searchFilter} onChange={(e) => setSearchFilter(e.target.value)}
                    className="py-3 px-4 pl-11 block w-full shadow-sm text-base outline-none text-gray-300 bg-stone-800"/>
                <div className="absolute inset-y-0 left-0 flex items-center pointer-events-none pl-4">
                    <svg className="h-4 w-4 text-gray-400" xmlns="http://www.w3.org/2000/svg" width="16" height="16"
                        fill="currentColor" viewBox="0 0 16 16">
                        <path d="M11.742 10.344a6.5 6.5 0 1 0-1.397 1.398h-.001c.03.04.062.078.098.115l3.85 3.85a1
                            1 0 0 0 1.415-1.414l-3.85-3.85a1.007 1.007 0 0 0-.115-.1zM12 6.5a5.5 5.5 0 1 1-11 0 5.5
                            5.5 0 0 1 11 0z"/>
                    </svg>
                </div>
            </div>

            {/* Pipeline list */}
            <div className="w-full overflow-y-auto flex flex-col scroll-smooth" ref={pipelinesContainer}>
                { pipelines?.filter(streamFilter).sort(streamSorter).map((record, key) => (
                    <div key={key} onClick={() => showDetails(record.name)} className="hover:bg-blue-600 flex
                        justify-between px-2 items-center hover:cursor-pointer">
                        <div className="flex items-center">
                            <div title={record.status}
                                className={"mr-2 w-4 h-4 rounded-full border-2 border-stone-400 " + record.status}>
                            </div>
                            <span className="py-2 px-1">{record.name}</span>
                        </div>
                        <div className="flex flex-col py-2 px-1 font-light text-indigo-300">
                            <span>{record.last_run ?? "-"}</span>
                            <span>{record.runtime ?? "--:--"}</span>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
};

export default PipelineListPanel;