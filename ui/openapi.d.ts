/**
 * This file was auto-generated by openapi-typescript.
 * Do not make direct changes to the file.
 */

export interface paths {
    "/login/clients": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Get login clients
         * @description Get available OAuth clients for login.
         */
        get: operations["getLoginClients"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/current-user": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Get current user
         * @description Get the currently logged in user
         */
        get: operations["getCurrentUser"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/organizations": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Get organizations
         * @description Get a list of organizations
         */
        get: operations["getOrganizations"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/organizations/{organization}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Get organization
         * @description Get an organization
         */
        get: operations["getOrganization"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
}
export type webhooks = Record<string, never>;
export interface components {
    schemas: {
        /** @description Login client */
        Client: {
            name?: string;
            logo?: string;
            requestPath?: string;
        };
        /** @description Organization */
        Organization: {
            id?: string;
            name?: string;
        };
        /** @description Workspace */
        Workspace: {
            id?: string;
            name?: string;
            organization?: string;
            terraformVersion?: string;
        };
        User: {
            username: string;
            isSiteAdmin?: boolean;
        };
        Pagination: {
            current_page: number;
            previous_page?: number;
            next_page?: number;
            total_pages: number;
            total_count: number;
        };
    };
    responses: {
        /** @description No content */
        EmptyOkResponse: {
            headers: {
                [name: string]: unknown;
            };
            content?: never;
        };
        /** @description Unauthorized */
        Unauthorized: {
            headers: {
                [name: string]: unknown;
            };
            content?: never;
        };
        /** @description Unexpected server error */
        InternalServerError: {
            headers: {
                [name: string]: unknown;
            };
            content: {
                "application/json": {
                    error: string;
                };
            };
        };
    };
    parameters: {
        /** @description Request ordinal page number. */
        page_number: number;
        /** @description Request max number of items on page. */
        page_size: number;
    };
    requestBodies: never;
    headers: never;
    pathItems: never;
}
export type $defs = Record<string, never>;
export interface operations {
    getLoginClients: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Client"][];
                };
            };
            500: components["responses"]["InternalServerError"];
        };
    };
    getCurrentUser: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
            };
            401: components["responses"]["Unauthorized"];
        };
    };
    getOrganizations: {
        parameters: {
            query?: {
                /** @description Request ordinal page number. */
                page_number?: components["parameters"]["page_number"];
                /** @description Request max number of items on page. */
                page_size?: components["parameters"]["page_size"];
            };
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": {
                        data?: components["schemas"]["Organization"][];
                        pagination?: components["schemas"]["Pagination"];
                    };
                };
            };
            401: components["responses"]["Unauthorized"];
            500: components["responses"]["InternalServerError"];
        };
    };
    getOrganization: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description Name of organization */
                organization: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
            };
            401: components["responses"]["Unauthorized"];
            500: components["responses"]["InternalServerError"];
        };
    };
}
