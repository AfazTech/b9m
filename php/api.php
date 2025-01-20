<?php

class DNSManager {
    private $apiUrl;

    public function __construct($apiUrl) {
        $this->apiUrl = rtrim($apiUrl, '/');
    }

    private function sendRequest($method, $endpoint, $data = null) {
        $url = $this->apiUrl . $endpoint;
        $ch = curl_init();

        curl_setopt($ch, CURLOPT_URL, $url);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_CUSTOMREQUEST, strtoupper($method));

        if ($data) {
            curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($data));
            curl_setopt($ch, CURLOPT_HTTPHEADER, [
                'Content-Type: application/json',
                'Content-Length: ' . strlen(json_encode($data))
            ]);
        }

        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        curl_close($ch);

        if ($httpCode >= 200 && $httpCode < 300) {
            return json_decode($response, true);
        } else {
            throw new Exception("Error: " . $response);
        }
    }

    public function addDomain($domain, $ns1, $ns2) {
        $data = [
            'domain' => $domain,
            'ns1' => $ns1,
            'ns2' => $ns2
        ];
        return $this->sendRequest('POST', '/domains', $data);
    }

    public function deleteDomain($domain) {
        return $this->sendRequest('DELETE', '/domains/' . $domain);
    }

    public function addRecord($domain, $name, $type, $value, $ttl) {
        $data = [
            'name' => $name,
            'type' => $type,
            'value' => $value,
            'ttl' => $ttl
        ];
        return $this->sendRequest('POST', '/domains/' . $domain . '/records', $data);
    }

    public function deleteRecord($domain, $name) {
        return $this->sendRequest('DELETE', '/domains/' . $domain . '/records/' . $name);
    }

    public function getAllRecords($domain) {
        return $this->sendRequest('GET', '/domains/' . $domain . '/records');
    }
}
