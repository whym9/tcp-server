CREATE DATABASE pcap_files;

USE pcap_files;

CREATE TABLE File (
    FileID int NOT NULL,
    FileName varchar(255),
    Size int,
    PRIMARY KEY (ID)
);

CREATE TABLE Packets (
    PacketID INT NOT NULL ,
    Timer TIME,
    Source VARCHAR(32),
    Destination VARCHAR(32),
    Protocol VARCHAR(16),
    Size INT,
    PRIMARY KEY (PacketID),
    Data VARBINARY(65535),
    FOREIGN KEY (FileID) REFERENCES File(FileID)
);

CREATE TABLE Statistics (
    ID INT NOT NULL,
    ProtocolTCP INT,
    UDP INT,
    IPv4 INT,
    IPv6 INT,
    PRIMARY KEY (ID),
    FOREIGN KEY (FileID) REFERENCES File(FileID)
);


