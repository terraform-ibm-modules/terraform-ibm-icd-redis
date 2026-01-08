#!/usr/bin/env python3
import http.client
import json
import os
import sys
from urllib.parse import urlparse


def parse_input():
    """
    Reads JSON input from stdin and parses it into a dictionary.
    Returns:
        dict: Parsed input data.
    """
    try:
        data = json.loads(sys.stdin.read())
    except json.JSONDecodeError as e:
        raise ValueError("Invalid JSON input") from e
    return data


def validate_inputs(data):
    """
    Validates required inputs 'IAM_TOKEN', 'REGION' and 'DB_TYPE' from the parsed input.
    Args:
        data (dict): Input data parsed from JSON.
    Returns:
        tuple: A tuple containing (IAM_TOKEN, REGION, DB_TYPE).
    """
    token = data.get("IAM_TOKEN")
    if not token:
        raise ValueError("IAM_TOKEN is required")

    region = data.get("REGION")
    if not region:
        raise ValueError("REGION is required")

    db_type = data.get("DB_TYPE")
    if not db_type:
        raise ValueError("DB_TYPE is required")

    return token, region, db_type


def get_api_endpoint(region):
    """
    Retrieves the API endpoint from environment variable or defaults to region-based URL.
    Args:
        region (str): Region to construct the default URL.
    Returns:
        str: The API endpoint URL.
    """
    api_endpoint = os.getenv("IBMCLOUD_ICD_API_ENDPOINT")
    if not api_endpoint:
        api_endpoint = f"https://api.{region}.databases.cloud.ibm.com"
    return api_endpoint


def fetch_icd_deployables(iam_token, api_endpoint):
    """
    Fetches ICD deployables versions using HTTP connection.
    Args:
        iam_token (str): IBM Cloud IAM token for authentication.
        api_endpoint (str): The API endpoint to use.
    Returns:
        dict: Parsed JSON response containing deployables information.
    """
    parsed = urlparse(api_endpoint)
    host = parsed.hostname
    
    # Remove 'Bearer ' prefix if present to avoid double prefixing
    if iam_token.startswith("Bearer "):
        iam_token = iam_token[7:]

    headers = {
        "Authorization": f"Bearer {iam_token}",
        "Accept": "application/json",
    }

    conn = http.client.HTTPSConnection(host)
    try:
        # Final API path
        url = "/v5/ibm/deployables"
        conn.request("GET", url, headers=headers)
        response = conn.getresponse()
        data = response.read().decode()

        if response.status != 200:
            raise RuntimeError(
                f"API request failed: {response.status} {response.reason} - {data}"
            )

        return json.loads(data)
    except http.client.HTTPException as e:
        raise RuntimeError("HTTP request failed") from e
    finally:
        conn.close()


def transform_data(deployables_data, db_type):
    """
    Extracts versions for the specific DB_TYPE.
    Args:
        deployables_data (dict): Raw data returned by the API.
        db_type (str): The type of database to filter for (e.g., 'redis').
    Returns:
        list: A list of version strings.
    """
    versions = []
    
    deployables = deployables_data.get("deployables", [])
    
    for item in deployables:
        if item.get("type") == db_type:
             for ver in item.get("versions", []):
                 if ver.get("status") not in ["dead", "hidden"]:
                     versions.append(ver.get("version"))
             # Found the db type, no need to continue unless there are duplicates which shouldn't happen
             break
             
    if not versions:
        # It's possible the DB_TYPE is valid but no versions found, or invalid DB_TYPE
        # For our purpose, if we don't find any versions, it might be an issue.
        # But we will return empty list and let terraform validation fail if it tries to match.
        pass

    return versions


def format_for_terraform(versions):
    """
    Converts the versions list into a JSON string for Terraform external data source consumption.
    Args:
        versions (list): List of version strings.
    Returns:
        dict: A dictionary with a single key 'versions' containing the JSON string of the list.
    """
    # Terraform external data source expects a flat map of strings.
    # So we encode the list as a JSON string.
    return {"versions": json.dumps(versions)}


def main():
    """
    Main execution function.
    """
    data = parse_input()
    iam_token, region, db_type = validate_inputs(data)
    
    api_endpoint = get_api_endpoint(region)
    deployables_data = fetch_icd_deployables(iam_token, api_endpoint)
    versions = transform_data(deployables_data, db_type)
    output = format_for_terraform(versions)

    print(json.dumps(output))


if __name__ == "__main__":
    main()
